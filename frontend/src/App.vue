<script setup lang="ts">
import { client } from './openapi/fetch'

// await client.POST('/token', {
//   body: { deviceID: 'web', email: 'test@test.it', password: 'password' },
// })
client.GET('/user')
</script>

<template>
  <h1>You did it!</h1>
  <p>
    Visit <a href="https://vuejs.org/" target="_blank" rel="noopener">vuejs.org</a> to read the
    documentation
  </p>
</template>

<style scoped></style>
