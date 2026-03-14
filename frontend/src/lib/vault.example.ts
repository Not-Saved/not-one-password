// ---------- Helpers ----------
function toBase64(buffer: ArrayBuffer | Uint8Array): string {
  if (buffer instanceof Uint8Array) {
    return btoa(String.fromCharCode(...buffer))
  }
  return btoa(String.fromCharCode(...new Uint8Array(buffer)))
}

function fromBase64(base64: string): BufferSource {
  return Uint8Array.from(atob(base64), (c) => c.charCodeAt(0))
}

// ---------- Key Derivation ----------
async function deriveKey(masterPassword: string, salt: BufferSource) {
  const encoder = new TextEncoder()
  const baseKey = await crypto.subtle.importKey(
    'raw',
    encoder.encode(masterPassword),
    'PBKDF2',
    false,
    ['deriveKey'],
  )
  return crypto.subtle.deriveKey(
    { name: 'PBKDF2', salt, iterations: 600_000, hash: 'SHA-256' },
    baseKey,
    { name: 'AES-GCM', length: 256 },
    false,
    ['encrypt', 'decrypt'],
  )
}

// ---------- Encryption / Decryption ----------
type EncryptedEntry = { iv: string; ciphertext: string }

async function encryptEntry(key: CryptoKey, entry: object): Promise<EncryptedEntry> {
  const encoder = new TextEncoder()
  const iv = crypto.getRandomValues(new Uint8Array(12))
  const ciphertext = await crypto.subtle.encrypt(
    { name: 'AES-GCM', iv },
    key,
    encoder.encode(JSON.stringify(entry)),
  )
  return { iv: toBase64(iv), ciphertext: toBase64(ciphertext) }
}

async function decryptEntry<T = unknown>(key: CryptoKey, encrypted: EncryptedEntry): Promise<T> {
  const decoder = new TextDecoder()
  const iv = fromBase64(encrypted.iv)
  const ciphertext = fromBase64(encrypted.ciphertext)
  try {
    const plaintextBuffer = await crypto.subtle.decrypt({ name: 'AES-GCM', iv }, key, ciphertext)
    return JSON.parse(decoder.decode(plaintextBuffer))
  } catch {
    throw new Error('Wrong master password or corrupted data')
  }
}

// ---------- Password Field Random Key Encryption ----------
type EncryptedPasswordField = {
  key: EncryptedEntry // password key encrypted with vault key
  iv: string // for AES-GCM of password
  ciphertext: string // encrypted password
}

async function encryptPasswordField(
  vaultKey: CryptoKey,
  password: string,
): Promise<EncryptedPasswordField> {
  // 1️⃣ Generate random password key
  const passwordKey = await crypto.subtle.generateKey({ name: 'AES-GCM', length: 256 }, true, [
    'encrypt',
    'decrypt',
  ])

  // 2️⃣ Encrypt password with passwordKey
  const encryptedPassword = await encryptEntry(passwordKey, { password })

  // 3️⃣ Export passwordKey to raw bytes
  const rawKey = await crypto.subtle.exportKey('raw', passwordKey)

  // 4️⃣ Encrypt passwordKey with vaultKey
  const encryptedKey = await encryptEntry(vaultKey, Array.from(new Uint8Array(rawKey)))

  return {
    key: encryptedKey,
    iv: encryptedPassword.iv,
    ciphertext: encryptedPassword.ciphertext,
  }
}

async function decryptPasswordField(
  vaultKey: CryptoKey,
  encrypted: EncryptedPasswordField,
): Promise<string> {
  // 1️⃣ Decrypt passwordKey with vaultKey
  const rawKeyArray: number[] = await decryptEntry<number[]>(vaultKey, encrypted.key)
  const rawKey = new Uint8Array(rawKeyArray)
  const passwordKey = await crypto.subtle.importKey('raw', rawKey, 'AES-GCM', true, ['decrypt'])

  // 2️⃣ Decrypt password with passwordKey
  const decrypted = await decryptEntry<{ password: string }>(passwordKey, {
    iv: encrypted.iv,
    ciphertext: encrypted.ciphertext,
  })
  return decrypted.password
}

// ---------- Vault Structure ----------
type VerificationBlock = EncryptedEntry
type DecryptedVerificationBlock = { check: string }
type VaultEntry = EncryptedEntry & { id: string; createdAt: number; updatedAt: number }
type DecryptedVaultEntry = {
  id: string
  createdAt: number
  updatedAt: number
  password: EncryptedPasswordField
  [key: string]: any
}

type Vault = {
  version: number
  kdf: { algorithm: string; hash: string; iterations: number; salt: string }
  verification: VerificationBlock
  entries: VaultEntry[]
  createdAt: number
  updatedAt: number
}

// ---------- Vault Operations ----------
export async function createVault(masterPassword: string): Promise<Vault> {
  const salt = crypto.getRandomValues(new Uint8Array(16))
  const key = await deriveKey(masterPassword, salt)
  const verification = await encryptEntry(key, { check: 'vault' })
  const now = Date.now()
  return {
    version: 1,
    kdf: { algorithm: 'PBKDF2', hash: 'SHA-256', iterations: 600_000, salt: toBase64(salt.buffer) },
    verification,
    entries: [],
    createdAt: now,
    updatedAt: now,
  }
}

export async function verifyMasterPassword(vault: Vault, masterPassword: string): Promise<boolean> {
  const salt = fromBase64(vault.kdf.salt)
  const key = await deriveKey(masterPassword, salt)
  try {
    const result = await decryptEntry<DecryptedVerificationBlock>(key, vault.verification)
    return result.check === 'vault'
  } catch {
    return false
  }
}

async function addEntry(
  vault: Vault,
  masterPassword: string,
  entryData: { password: string; [key: string]: any },
) {
  const salt = fromBase64(vault.kdf.salt)
  const vaultKey = await deriveKey(masterPassword, salt)

  // Encrypt password field with random key
  const encryptedPassword = await encryptPasswordField(vaultKey, entryData.password)

  // Replace password with encrypted object
  const entryWithEncryptedPassword = { ...entryData, password: encryptedPassword }

  // Encrypt full entry
  const encryptedEntry = await encryptEntry(vaultKey, entryWithEncryptedPassword)

  const now = Date.now()
  vault.entries.push({ id: crypto.randomUUID(), ...encryptedEntry, createdAt: now, updatedAt: now })
  vault.updatedAt = now
}

async function decryptVaultEntry(
  vault: Vault,
  masterPassword: string,
  entryId: string,
): Promise<DecryptedVaultEntry> {
  const salt = fromBase64(vault.kdf.salt)
  const vaultKey = await deriveKey(masterPassword, salt)

  const entry = vault.entries.find((e) => e.id === entryId)
  if (!entry) throw new Error('Entry not found')

  const decrypted = await decryptEntry<DecryptedVaultEntry>(vaultKey, entry)
  return decrypted
}

async function decryptPassword(vaultKey: CryptoKey, encryptedPassword: EncryptedPasswordField) {
  return decryptPasswordField(vaultKey, encryptedPassword)
}

async function rotateMasterPassword(
  vault: Vault,
  oldMasterPassword: string,
  newMasterPassword: string,
) {
  // 1️⃣ Derive old vault key
  const oldSalt = fromBase64(vault.kdf.salt)
  const oldVaultKey = await deriveKey(oldMasterPassword, oldSalt)

  // 2️⃣ Generate new salt and new vault key
  const newSalt = crypto.getRandomValues(new Uint8Array(16))
  const newVaultKey = await deriveKey(newMasterPassword, newSalt)

  // 3️⃣ Re-encrypt verification block
  const verificationPlain = await decryptEntry<VerificationBlock>(oldVaultKey, vault.verification)
  const newVerification = await encryptEntry(newVaultKey, verificationPlain)

  // 4️⃣ Re-encrypt all entries
  const newEntries: VaultEntry[] = []
  for (const entry of vault.entries) {
    // Decrypt full entry
    const decryptedEntry = await decryptEntry<DecryptedVaultEntry>(oldVaultKey, entry)

    // Decrypt the password field
    const passwordPlain = await decryptPassword(oldVaultKey, decryptedEntry.password)

    // Re-encrypt password field with new vault key
    const newEncryptedPassword = await encryptPasswordField(newVaultKey, passwordPlain)

    // Replace password field
    decryptedEntry.password = newEncryptedPassword

    // Re-encrypt full entry with new vault key
    const reEncryptedEntry = await encryptEntry(newVaultKey, decryptedEntry)

    newEntries.push({
      id: entry.id,
      createdAt: entry.createdAt,
      updatedAt: Date.now(),
      ...reEncryptedEntry,
    })
  }

  // 5️⃣ Update vault metadata
  vault.kdf.salt = toBase64(newSalt)
  vault.kdf.iterations = 600_000 // keep PBKDF2 params same (or change if desired)
  vault.verification = newVerification
  vault.entries = newEntries
  vault.updatedAt = Date.now()
}
// ---------- Demo ----------
async function demo() {
  const masterPassword = 'correct horse battery staple'

  const vault = await createVault(masterPassword)
  console.log('Vault created:', vault)

  const isValid = await verifyMasterPassword(vault, masterPassword)
  console.log('Master password valid?', isValid)

  await addEntry(vault, masterPassword, {
    site: 'github.com',
    username: 'alice',
    password: 'mypassword123',
  })
  console.log('Vault after adding entry:', vault)

  const entryId = vault.entries[0]!.id
  const vaultKey = await deriveKey(masterPassword, fromBase64(vault.kdf.salt))
  const entry = await decryptVaultEntry(vault, masterPassword, entryId)
  console.log('Vault entry without password:', entry)

  const password = await decryptPassword(vaultKey, entry.password)
  console.log('Decrypted password:', password)

  await rotateMasterPassword(vault, masterPassword, 'new stronger password')
  console.log('Vault after rotation:', vault)

  const isValidAfter = await verifyMasterPassword(vault, 'new stronger password')
  console.log('New password valid?', isValidAfter)

  const compressedVault = await compressVault(vault)
  console.log(compressedVault)

  const decompressedVault = await decompressVault(compressedVault)
  console.log(decompressedVault)
}

export async function compressVault(vault: object): Promise<Uint8Array> {
  const vaultString = JSON.stringify(vault)
  const encoder = new TextEncoder()
  const vaultBytes = encoder.encode(vaultString)

  // 1️⃣ Create a ReadableStream from the Uint8Array
  const readable = new ReadableStream({
    start(controller) {
      controller.enqueue(vaultBytes)
      controller.close()
    },
  })

  // 2️⃣ Pipe through CompressionStream
  const cs = new CompressionStream('gzip')
  const compressedStream = readable.pipeThrough(cs)

  // 3️⃣ Collect compressed chunks
  const reader = compressedStream.getReader()
  const chunks: Uint8Array[] = []
  let done = false
  while (!done) {
    const { value, done: readerDone } = await reader.read()
    done = readerDone
    if (value) chunks.push(value)
  }

  // 4️⃣ Concatenate into single Uint8Array
  const size = chunks.reduce((acc, c) => acc + c.length, 0)
  const result = new Uint8Array(size)
  let offset = 0
  for (const chunk of chunks) {
    result.set(chunk, offset)
    offset += chunk.length
  }

  return result
}

// Decompress a GZIP-compressed Uint8Array in the browser
export async function decompressVault(compressed: Uint8Array): Promise<Vault> {
  // 1️⃣ Create a ReadableStream from the compressed bytes
  const readable = new ReadableStream({
    start(controller) {
      controller.enqueue(compressed)
      controller.close()
    },
  })

  // 2️⃣ Pipe through DecompressionStream
  const ds = new DecompressionStream('gzip')
  const decompressedStream = readable.pipeThrough(ds)

  // 3️⃣ Collect chunks
  const reader = decompressedStream.getReader()
  const chunks: Uint8Array[] = []
  let done = false
  while (!done) {
    const { value, done: readerDone } = await reader.read()
    done = readerDone
    if (value) chunks.push(value)
  }

  // 4️⃣ Concatenate chunks into a single Uint8Array
  const totalSize = chunks.reduce((sum, c) => sum + c.length, 0)
  const result = new Uint8Array(totalSize)
  let offset = 0
  for (const chunk of chunks) {
    result.set(chunk, offset)
    offset += chunk.length
  }

  // 5️⃣ Decode to string and parse JSON
  const decoder = new TextDecoder()
  const jsonString = decoder.decode(result)
  return JSON.parse(jsonString)
}

async function compareVaultCompression(vault: object) {
  const vaultString = JSON.stringify(vault)
  const encoder = new TextEncoder()
  const vaultBytes = encoder.encode(vaultString)
  console.log('Original size (bytes):', vaultBytes.length)

  const compressed = await compressVault(vault)
  console.log('Compressed size (bytes):', compressed.length)

  const ratio = (compressed.length / vaultBytes.length).toFixed(2)
  console.log('Compression ratio:', ratio)

  return { original: vaultBytes.length, compressed: compressed.length, ratio }
}
