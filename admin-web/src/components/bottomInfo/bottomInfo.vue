<template>
  <div
    class="flex flex-col md:flex-row gap-2 items-center text-sm text-slate-700 dark:text-slate-500 justify-center py-2"
  >
    <div class="text-center font-bold text-active">MeAdmin</div>
    <slot />
  </div>
</template>

<script setup>
  defineOptions({
    name: 'BottomInfo'
  })

  console.log('%c MeAdmin ', 'background:#354855; padding: 1px 5px; border-radius: 3px; color: #fff; font-weight: bold;')
</script>
