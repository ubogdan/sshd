<template>
  <el-card sytle="mini-height:100%" v-loading="loading">
    <el-row slot="header" type="flex" justify="space-between" align="middle">
      <el-col :span="18">
        <el-breadcrumb separator-class="el-icon-arrow-right">
          <el-breadcrumb-item v-for="(name,index) in breads" :key="name">
            <el-link @click="doBreadcrumb(index)" v-text="name" :disabled="breads.length === (index+1)"
                     type="primary"></el-link>
          </el-breadcrumb-item>
        </el-breadcrumb>
      </el-col>
      <el-col :span="6" style="display: flex;justify-content: flex-end;align-items: center ">
        <el-input v-model="search" clearable placeholder="search file or a directory" size="mini" style="margin-right: 10px"></el-input>
        <el-button size="mini" title="search a file or a dir" @click="doSearch" icon="el-icon-search"></el-button>
        <el-button type="success" icon="el-icon-upload" size="mini" title="upload a file" @click="doUpload"></el-button>
        <el-button type="primary" icon="el-icon-folder-add" size="mini" title="make a folder" @click="doMakeDir"></el-button>
        <el-button type="warning" icon="el-icon-refresh" size="mini" title="refresh" @click="doRefresh"></el-button>
      </el-col>
    </el-row>

    <el-table
        :data="items"
        stripe
        style="width: 100%">
      <el-table-column
          label="i"
          width="180">
        <template slot-scope="scope">
          <el-link type="primary" plain :icon="scope.row.is_dir ? 'el-icon-folder-opened' : 'el-icon-document-checked'" @click="openDir(scope.row)" :disabled="!scope.row.is_dir"></el-link>
        </template>
      </el-table-column>

      <el-table-column
          prop="name"
          label="名称"
          width="180">
      </el-table-column>
      <el-table-column
          prop="mod"
          label="mode"
          width="180">
      </el-table-column>
      <el-table-column
          prop="size"
          label="大小">
      </el-table-column>
      <el-table-column
          prop="time"
          label="时间">
      </el-table-column>

      <el-table-column label="操作">
        <template slot-scope="scope">
          <el-button size="mini" icon="el-icon-download" @click="doDownload(scope.row.path,scope.row.is_dir)" type="success"></el-button>
          <el-button size="mini" icon="el-icon-delete-solid" @click="doRm(scope.row)" type="danger"></el-button>
        </template>
      </el-table-column>
    </el-table>

    <df-upload :dir="dir" :token="token" :visible="v" @afterClose="v = false;doRefresh()"></df-upload>


  </el-card>


</template>

<script>
import DfUpload from "../components/DfUpload";

export default {
  name: 'WebSftp',
  components: {DfUpload},
  props: {
    token: String,
    msg: String
  },
  data: function () {
    return {
      v: false,
      loading: false,
      search: '',
      dir: '',
      visible: false,
      items: [],
      breads: [],
    }
  },
  computed: {},
  beforeUpdate() {
    window.document.title = "RemoteDesktop"
  },
  mounted() {
    this.dir = this.$route.query.dir
    this.apiSftpLs(this.dir)
  },
  methods: {
    doUpload() {
      this.v = true
    },
    doRefresh() {
      this.apiSftpLs(this.dir)
    },
    doMakeDir() {
      this.$prompt('请输入要创建的目录', '创建目录', {
        confirmButtonText: '确定',
        cancelButtonText: '取消',
        // inputPattern: /[\w!#$%&'*+/=?^_`{|}~-]+(?:\.[\w!#$%&'*+/=?^_`{|}~-]+)*@(?:[\w](?:[\w-]*[\w])?\.)+[\w](?:[\w-]*[\w])?/,
        // inputErrorMessage: '邮箱格式不正确'
      }).then(({value}) => {
        this._apiMkdir(value)
      }).catch(() => {
        this.$message({
          type: 'info',
          message: '取消输入'
        });
      });


    },
    _apiMkdir(dirName) {
      let path = this.dir + "/" + dirName;
      let token = this.token;
      this.loading = true
      this.$http.get('/api/sftp/mkdir', {params: {token, path}}).then(() => {
        this.loading = false
        this.doRefresh()
      }).catch(e => {
        this.loading = false;
        this.$message.error(e)
      })
    },
    doSearch() {
      this.apiSftpLs(this.dir)
    },
    doDownload(path, isDir) {
      let url = `/api/sftp/download/${isDir ? 'dir' : 'file'}?token=${this.token}&path=${encodeURIComponent(path)}`
      let win = window.open(url, '_blank');
      win.focus();
    },
    doRm(item) {
      let token = this.token;
      let path = item.path;
      let dir_or_file = item.is_dir ? 'dir' : 'file'
      this.$http.get('/api/sftp/rm', {params: {token, path, dir_or_file}}).then(res => {
        this.loading = false
        this.apiSftpLs(this.dir)
      }).catch(e => {
        this.loading = false
        this.$message.error(JSON.stringify(e));
      })
    },
    openDir(item) {
      // let url =  `?dir=${encodeURIComponent(item.path)}&token=${this.token}`
      // window.location = url
      this.apiSftpLs(item.path)
    },
    doBreadcrumb(idx) {
      let elements = []
      for (let i = 0; i <= idx; i++) {
        elements.push(this.breads[i])
      }
      let dir = elements.join('/').replace('//', '/')
      this.apiSftpLs(dir)
    },
    _setBreads(dir) {
      this.dir = dir

      //todo:: there is a bug
      if (dir === '/') {
        this.breads = ['/']
        return
      }
      let list = dir.split('/')
      for (let i = 0; i < list.length; i++) {
        if (list[i] === "") {
          list[i] = '/'
        }
      }
      this.breads = list
    },
    apiSftpLs(dir) {
      this.loading = true
      let token = this.token
      let search = this.search

      this.$http.get('/api/sftp/ls', {params: {token, dir, search}}).then(({list, dir}) => {
            this.loading = false
            this.items = list
            this._setBreads(dir)
          }
      ).catch(e => {
        this.loading = false
        this.$message.error(JSON.stringify(e));
      })
    },
  }
}


</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
h3 {
  margin: 40px 0 0;
}

ul {
  list-style-type: none;
  padding: 0;
}

li {
  display: inline-block;
  margin: 0 10px;
}

a {
  color: #42b983;
}
</style>
