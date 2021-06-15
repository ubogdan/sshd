

<template>

  <el-dialog
      :title="title"
      :visible.sync="visible"
      @close="$emit('afterClose')"
      @open="doOpen"
  >
    <el-upload
        ref="upload"
        :action="action"
        :on-preview="handlePreview"
        :on-remove="handleRemove"
        :file-list="fileList"
        :data="form"
        :auto-upload="false">
      <el-button slot="trigger" size="small" type="primary">选取文件</el-button>
      <el-button style="margin-left: 10px;" size="small" type="success" @click="submitUpload">上传到服务器</el-button>
      <div slot="tip" class="el-upload__tip">只能上传jpg/png文件，且不超过500kb</div>
    </el-upload>

  </el-dialog>

</template>

<script>


export default {
  props: {visible: Boolean,token:String,dir: String},
  name: "DfUpload",
  data() {

    return {
      action:'/api/sftp/upload',
      title:'上传文件',
      fileList: [],
      form:{},
    };
  },
  methods: {

    doOpen() {
      this.form.token = this.token
      this.form.dir = this.dir
    },
    submitUpload() {

      this.$refs.upload.submit();
      this.$emit("afterClose")
    },
    handleRemove(file, fileList) {
      console.log(file, fileList);
    },
    handlePreview(file) {
      console.log(file);
    }
  },
};
</script>
<style scoped>

</style>
