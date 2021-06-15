<template>
  <div class="swagger" id="swagger"></div>
</template>

<script>
import SwaggerUI from 'swagger-ui';
import 'swagger-ui/dist/swagger-ui.css';
export default {
  name: "Swagger",
  beforeUpdate() {
    window.document.title = "SwaggerDoc"
  },
  mounted() {
    window.document.title = "RemoteDesktop"
    const spec = require('../../../docs/swagger.json');
    SwaggerUI({
      spec: spec,
      dom_id: '#swagger'
    })
  }
}
</script>

<style scoped>

</style>