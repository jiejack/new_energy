<template>
  <div id="app">
    <router-view />
    <GlobalNotification />
  </div>
</template>

<script setup lang="ts">
import GlobalNotification from '@/components/GlobalNotification/index.vue'
</script>

<style>
/* Global styles are imported in main.ts */
#app {
  width: 100%;
  height: 100vh;
  overflow: hidden;
}
</style>
