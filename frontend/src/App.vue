<script setup lang="ts">
import { computed, onMounted, ref } from 'vue'
import { client, clientWithoutRefresh } from './openapi/fetch'
import { type SchemaUserResponse } from './openapi/openapi'
import {
  addEntry,
  compressVault,
  createVault,
  decompressVault,
  decryptVault,
  type DecryptedVault,
} from './lib/vault.example'

async function login() {
  const res = await client.POST('/token', {
    body: {
      deviceID: 'web',
      email: 'test@test.it',
      password: 'password',
    },
  })
  if (res.response.ok) {
    getUser()
  } else {
    removeUser()
  }
}

const user = ref<SchemaUserResponse | null>(null)
async function getUser() {
  const res = await client.GET('/user')
  if (res.data) {
    user.value = res.data
  }
}

function removeUser() {
  user.value = null
}

async function logout() {
  await client.POST('/logout')
  removeUser()
  vault.value = null
}

async function refresh() {
  await clientWithoutRefresh.POST('/refresh')
}

async function initVault() {
  const vault = await createVault('my password')
  await addEntry(vault, 'my password', {
    id: '1',
    name: 'Google',
    username: 'test',
    password: 'password',
  })
  const vaultCompressed = await compressVault(vault)
  const res = await client.POST('/user/vault', {
    body: '',
    headers: {
      'content-type': 'application/octet-stream',
    },
    bodySerializer: () => {
      return vaultCompressed
    },
  })
  return res
}

const registration = ref({
  email: '',
  password: '',
  name: '',
})
async function register() {
  await client.POST('/user', {
    body: registration.value,
  })
}

const confirmUserCode = ref('')
async function confirmUser() {
  const res = await client.POST('/user/confirm', {
    params: {
      query: {
        code: confirmUserCode.value,
      },
    },
  })
}

const vault = ref<null | DecryptedVault>(null)
async function getVault() {
  const res = await client.GET('/user/vault', { parseAs: 'blob' })
  const blob = res.data as Blob

  const contentType = res.response.headers.get('content-type')!
  if (!contentType.startsWith('multipart/form-data')) {
    throw new Error('Cannot parse: not multipart/form-data')
  }

  const formData = await new Response(blob, {
    headers: { 'Content-Type': contentType },
  }).formData()

  const v: Record<string, any> = {}
  for (const [key, value] of formData.entries()) {
    if (value instanceof File) {
      const arrayBuffer = await value.arrayBuffer()
      const bufferSource = new Uint8Array(arrayBuffer)

      const ev = await decompressVault(bufferSource)
      const dv = await decryptVault(ev, 'my password')
      v['vault'] = dv
    } else {
      v[key] = value
    }
  }
  vault.value = v as DecryptedVault
}

onMounted(() => {
  getUser()
})

function isLogged() {
  return user.value !== null
}
</script>

<template>
  <h1>You did it!</h1>
  <p>
    Visit <a href="https://vuejs.org/" target="_blank" rel="noopener">vuejs.org</a> to read the
    documentation
  </p>
  <div>Logged: {{ user?.email }}</div>
  <button @click="login">Login</button>
  <button v-if="isLogged()" @click="logout">Logout</button>
  <button v-if="isLogged()" @click="refresh">Refresh</button>
  <button v-if="isLogged()" @click="initVault">Init vault</button>
  <button v-if="isLogged()" @click="getVault">Get vault</button>
  <div v-if="vault !== null" class="print">{{ JSON.stringify(vault, undefined, 4) }}</div>
  <div v-if="user === null">
    <form>
      <input v-model="registration.email" placeholder="Email" />
      <input v-model="registration.password" type="password" placeholder="Password" />
      <input v-model="registration.name" placeholder="Name" />
      <button @click="register">Register</button>
    </form>
  </div>
  <div v-if="user === null">
    <form>
      <input v-model="confirmUserCode" placeholder="Confirmation code" />
      <button @click="confirmUser">Confirm user</button>
    </form>
  </div>
</template>

<style scoped>
.print {
  white-space: pre-wrap;
}
input {
  display: block;
}
</style>
