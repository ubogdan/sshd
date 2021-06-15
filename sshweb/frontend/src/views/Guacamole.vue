<template>
  <el-container style="height: 100vh;width: 100%">
      <div ref="viewport" class="viewport">
        <!-- tabindex allows for div to be focused -->
        <div ref="display" class="display" tabindex="0"/>
      </div>

  </el-container>


</template>

<script>
import Guacamole from 'guacamole-common-js'
import GuacMouse from '../libs/GuacMouse'
import states from '../libs/states'
import clipboard from '../libs/clipboard'

Guacamole.Mouse = GuacMouse.mouse

function serialize(obj) {
  let str = [];
  for (const p in obj)
    if (obj[p]) {
      str.push(encodeURIComponent(p) + "=" + encodeURIComponent(obj[p]));
    }
  return str.join("&");
}

export default {
  props: {token: String},
  name: 'Guacamole',
  data() {
    return {
      demo: 'win',
      connected: false,
      display: null,
      currentAdjustedHeight: null,
      client: null,
      keyboard: null,
      mouse: null,
      lastEvent: null,
      connectionState: states.IDLE,
      errorMessage: '',
      arguments: {},
      query: {
        token: this.token,
        screen_width: 1024,
        screen_height: 768,
      },
    }
  },
  computed: {

    wsUrl() {
      let loc = window.location, new_uri;
      if (loc.protocol === "https:") {
        new_uri = "wss:";
      } else {
        new_uri = "ws:";
      }
      // new_uri += "//127.0.0.1:9528"
      new_uri += "//" + loc.host;
      new_uri += `/api/ws/guacamole`;
      return `${new_uri}`
    }
  },
  watch: {
    connectionState(state) {
      this.$notify({
        title: state,
        type: 'success'
      });
    }
  },

  methods: {
    doGuacdConnect() {
      window.document.title = "RemoteDesktop"
      this._setScreenSize()
      this.startGuacamole()
    },
    send(cmd) {
      if (!this.client) {
        return
      }
      for (const c of cmd.data) {
        this.client.sendKeyEvent(1, c.charCodeAt(0))
      }
    },
    copy(cmd) {
      if (!this.client) {
        return
      }
      clipboard.cache = {
        type: 'text/plain',
        data: cmd.data
      }
      clipboard.setRemoteClipboard(this.client)
    },
    handleMouseState(mouseState) {
      const scaledMouseState = Object.assign({}, mouseState, {
        x: mouseState.x / this.display.getScale(),
        y: mouseState.y / this.display.getScale(),
      })
      this.client.sendMouseState(scaledMouseState);
    },
    _setScreenSize() {
      const elm = this.$refs.viewport
      if (!elm || !elm.offsetWidth) {
        // resize is being called on the hidden window
        return
      }
      let pixelDensity = window.devicePixelRatio || 1
      const width = elm.clientWidth * pixelDensity
      const height = elm.clientHeight * pixelDensity
      this.query.screen_height = height
      this.query.screen_width = width
    },
    resize() {
      const elm = this.$refs.viewport
      if (!elm || !elm.offsetWidth) {
        // resize is being called on the hidden window
        return
      }

      let pixelDensity = window.devicePixelRatio || 1
      const width = elm.clientWidth * pixelDensity
      const height = elm.clientHeight * pixelDensity
      if (this.display.getWidth() !== width || this.display.getHeight() !== height) {
        this.client.sendSize(width, height)
      }
      // setting timeout so display has time to get the correct size
      setTimeout(() => {
        const scale = Math.min(
            elm.clientWidth / Math.max(this.display.getWidth(), 1),
            elm.clientHeight / Math.max(this.display.getHeight(), 1)
        )
        this.display.scale(scale)
      }, 100)
    },
    startGuacamole() {
      // GuacadAddr    string `form:"guacad_addr"`
      // AssetProtocol string `form:"asset_protocol"`
      // AssetHost     string `form:"asset_host"`
      // AssetPort     string `form:"asset_port"`
      // AssetUser     string `form:"asset_user"`
      // AssetPassword string `form:"asset_password"`
      // ScreenWidth   int    `form:"screen_width"`
      // ScreenHeight  int    `form:"screen_height"`
      // ScreenDpi     int    `form:"screen_dpi"`

      let tunnel = new Guacamole.WebSocketTunnel(this.wsUrl)


      if (this.client) {
        this.display.scale(0)
        this.uninstallKeyboard()
      }

      this.client = new Guacamole.Client(tunnel);
      clipboard.install(this.client)

      tunnel.onerror = status => {
        // eslint-disable-next-line no-console
        console.error(`Tunnel failed ${JSON.stringify(status)}`)
        this.connectionState = states.TUNNEL_ERROR
      }

      tunnel.onstatechange = state => {
        switch (state) {
            // Connection is being established
          case Guacamole.Tunnel.State.CONNECTING:
            this.connectionState = states.CONNECTING
            break;

            // Connection is established / no longer unstable
          case Guacamole.Tunnel.State.OPEN:
            this.connectionState = states.CONNECTED
            break;

            // Connection is established but misbehaving
          case Guacamole.Tunnel.State.UNSTABLE:
            // TODO
            //this.$message.error("不稳定")
            break;

            // Connection has closed
          case Guacamole.Tunnel.State.CLOSED:
            this.connectionState = states.DISCONNECTED
            break;
        }
      }

      this.client.onstatechange = clientState => {
        switch (clientState) {
          case 0:
            this.connectionState = states.IDLE
            break
          case 1:
            // connecting ignored for some reason?
            break
          case 2:
            this.connectionState = states.WAITING
            break
          case 3:
            this.connectionState = states.CONNECTED
            window.addEventListener('resize', this.resize)
            this.$refs.viewport.addEventListener('mouseenter', this.resize)

            clipboard.setRemoteClipboard(this.client)

            // eslint-disable-next-line no-fallthrough
          case 4:
          case 5:
            // disconnected, disconnecting
            break
        }
      }

      this.client.onerror = error => {
        this.client.disconnect()
        // eslint-disable-next-line no-console
        this.$message.error(`Client error ${JSON.stringify(error)}`)
        this.errorMessage = error.message
        this.connectionState = states.CLIENT_ERROR
      }

      this.client.onsync = () => {
      }

      // Test for argument mutability whenever an argument value is received
      this.client.onargv = (stream, mimetype, name) => {
        if (mimetype !== 'text/plain')
          return;

        const reader = new Guacamole.StringReader(stream);

        // Assemble received data into a single string
        let value = '';
        reader.ontext = text => {
          value += text;
        };

        // Test mutability once stream is finished, storing the current value for the argument only if it is mutable
        reader.onend = () => {
          const stream = this.client.createArgumentValueStream('text/plain', name)
          stream.onack = status => {
            if (status.isError()) {
              // ignore reject
              return
            }
            this.arguments[name] = value
          }
        };
      }

      this.client.onclipboard = clipboard.onClipboard
      this.display = this.client.getDisplay()
      const displayElm = this.$refs.display
      displayElm.appendChild(this.display.getElement());
      displayElm.addEventListener('contextmenu', e => {
        e.stopPropagation();
        if (e.preventDefault) {
          e.preventDefault();
        }
        e.returnValue = false;
      })

      let param = serialize(this.query)
      this.client.connect(param);
      window.onunload = () => this.client.disconnect();

      this.mouse = new Guacamole.Mouse(displayElm);
      // Hide software cursor when mouse leaves display
      this.mouse.onmouseout = () => {
        if (!this.display) return;
        this.display.showCursor(false);
      }

      // allows focusing on the display div so that keyboard doesn't always go to session
      displayElm.onclick = () => {
        displayElm.focus()
      }
      displayElm.onfocus = () => {
        displayElm.className = 'focus'
      }
      displayElm.onblur = () => {
        displayElm.className = ''
      }

      this.keyboard = new Guacamole.Keyboard(displayElm);
      this.installKeyboard()
      this.mouse.onmousedown = this.mouse.onmouseup = this.mouse.onmousemove = this.handleMouseState
      setTimeout(() => {
        this.resize()
        displayElm.focus()
      }, 1000); // $nextTick wasn't enough
    },
    installKeyboard() {
      this.keyboard.onkeydown = keysym => {
        this.client.sendKeyEvent(1, keysym);
      };
      this.keyboard.onkeyup = keysym => {
        this.client.sendKeyEvent(0, keysym);
      };
    },
    uninstallKeyboard() {
      this.keyboard.onkeydown = this.keyboard.onkeyup = () => {
      }
    }
  },
  beforeUpdate() {
    window.document.title = "RemoteDesktop"
  },
  mounted() {
    this._setScreenSize()
    this.doGuacdConnect()
  }
}
</script>

<style scoped>



.display {
  overflow: hidden;
  width: 100%;
  height: 100%;
}

.viewport {
  margin-top:10px;
  background-color: coral;
  border-color: blue;
  border-width: 1px;
  position: relative;
  width: 100%;
  height: 100%;
}
</style>
