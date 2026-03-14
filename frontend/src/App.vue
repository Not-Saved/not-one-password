<script setup lang="ts">
import { onMounted, ref } from 'vue'
import { client, clientWithoutRefresh } from './openapi/fetch'
import { type SchemaUserResponse } from './openapi/openapi'
import {
  compressVault,
  createVault,
  decompressVault,
  verifyMasterPassword,
} from './lib/vault.example'

// await client.POST('/token', {
//   body: { deviceID: 'web', email: 'test@test.it', password: 'password' },
// })
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
}

async function refresh() {
  await clientWithoutRefresh.POST('/refresh')
}
async function initVault() {
  const vault = await createVault('my password')
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

async function getVault() {
  const res = await client.GET('/user/vault', { parseAs: 'blob' })
  // 1️⃣ get blob
  const blob = res.data!

  // 2️⃣ only works if response is multipart/form-data
  const contentType = res.response.headers.get('content-type')!
  if (!contentType.startsWith('multipart/form-data')) {
    throw new Error('Cannot parse: not multipart/form-data')
  }

  // 3️⃣ wrap blob in Response and parse
  const formData = await new Response(blob, {
    headers: { 'Content-Type': contentType },
  }).formData()

  // 4️⃣ read parts
  for (const value of formData.values()) {
    if (value instanceof File) {
      // Convert File to ArrayBuffer
      const arrayBuffer = await value.arrayBuffer()
      // Convert to Uint8Array if you want a BufferSource
      const bufferSource = new Uint8Array(arrayBuffer)

      const vault = await decompressVault(bufferSource)
      const check = await verifyMasterPassword(vault, 'my password')
      console.log(check)
    }
  }
}

onMounted(() => {
  getUser()
})
</script>

<template>
  <h1>You did it!</h1>
  <p>
    Visit <a href="https://vuejs.org/" target="_blank" rel="noopener">vuejs.org</a> to read the
    documentation
  </p>
  <div>Logged: {{ user?.email }}</div>
  <button @click="login">Login</button>
  <button @click="logout">Logout</button>
  <button @click="refresh">Refresh</button>
  <button @click="initVault">Init vault</button>
  <button @click="getVault">Get vault</button>
</template>

<style scoped></style>
