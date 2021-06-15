<template>
  <section>
    <div id="terminal">


    </div>
    <el-button type="primary" size="small" icon="el-icon-connection" @click="doSFTP" style="position: absolute; top:2rem;right: 2rem;z-index:999999"></el-button>
  </section>

</template>

<script>

import {Terminal} from 'xterm';
import {FitAddon} from 'xterm-addon-fit';
import "xterm/css/xterm.css"


const MsgOperation_Stdin = 0
const MsgOperation_Stdout = 1
const MsgOperation_Resize = 2
const MsgOperation_Ping = 3

function websocketURLPrefix() {
  let loc = window.location, new_uri;
  if (loc.protocol === "https:") {
    new_uri = "wss:";
  } else {
    new_uri = "ws:";
  }
  new_uri += "//" + loc.host;
  new_uri += "/api/ws/";
  return new_uri
}

export default {
  name: "Ssh",
  props: {
    token: String,
  },
  data() {
    return {
      term: null,
      fitAddon: null,
      conn: null,
    }
  },

  mounted() {
    sessionStorage.setItem("token", this.token)
    this.init()
    this.start()
  },
  beforeUpdate() {
    window.document.title = "Terminal"
  },
  updated() {
    this.fitAddon.fit()
  },
  beforeDestroy() {
    this.close()
  },
  methods: {
    doSFTP(){
      let routeData = this.$router.resolve({name: 'WebSftp', params: {token:this.token}});
      window.open(routeData.href, '_blank');

    },
    init() {
      const term = new Terminal({
        fontSize: 16,
        cursorBlink: true,
        cursorStyle: 'bar',
        bellStyle: "sound",
      });
      const fitAddon = new FitAddon();

      term.loadAddon(fitAddon)
      term.open(document.getElementById("terminal"));
      term.write("connecting ...")
      fitAddon.fit();
      term.focus();
      this.term = term;
      this.fitAddon = fitAddon;
    },
    start() {
      let url = websocketURLPrefix()
      url = url + 'ssh' + `?t=${this.token}&r=${this.term.rows}&c=${this.term.cols}`
      const conn = new WebSocket(url);

      // term.toggleFullScreen(true);
      this.term.onData(data => {
        const msg = {operation: MsgOperation_Stdin, data: data}
        conn.send(JSON.stringify(msg))
      });
      this.term.onResize(size => {
        console.log("resize: " + size)
        if (conn.readyState === 1) {
          const msg = {operation: MsgOperation_Resize, cols: size.cols, rows: size.rows}
          conn.send(JSON.stringify(msg))
        }
      });


      //4. send ws heart beat
      this.timer = setInterval(() => {
        conn.send(JSON.stringify({operation: MsgOperation_Ping}));
      }, 15 * 1000);


      conn.onopen = (e) => {
        const msg = {operation: MsgOperation_Stdin, data: "export TERM=xterm && clear;\r"}
        conn.send(JSON.stringify(msg))
        // term.clear()

        const size = {operation: MsgOperation_Resize, cols: this.term.cols, rows: this.term.rows}
        conn.send(JSON.stringify(size))
      };
      conn.onmessage = (event) => {
        const msg = JSON.parse(event.data)
        if (msg.operation === MsgOperation_Stdout) {
          this.term.write(msg.data)
        } else {
          console.log("invalid msg operation: " + msg)
        }
      };
      conn.onclose = (event) => {
        if (event.wasClean) {
          console.log(`[close] Connection closed cleanly, code=${event.code} reason=${event.reason}`);
        } else {
          console.log('[close] Connection died');
          this.term.writeln("")
        }
        this.term.write('Connection Reset By Peer! Try Refresh.');
      };
      conn.onerror = (error) => {
        console.log('[error] Connection error');
        this.term.write("error: " + error.message);
        this.term.dispose();
      };
      this.fitAddon.fit()

      this.conn = conn
      //7. watch window size change
      window.onresize = () => {
        this.fitAddon.fit()
      };

    },
    close() {
      this.conn.close();
      clearInterval(this.timer)
      this.term.dispose()
    },

  }
}


</script>

<style scoped>
#terminal {
  height: 100vh;
  width: 100vw;
}

</style>