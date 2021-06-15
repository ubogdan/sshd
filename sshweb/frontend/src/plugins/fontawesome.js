import Vue from 'vue'

import { library } from '@fortawesome/fontawesome-svg-core'
import {faSearch,faUpload,faTrash,faFile,faFolder,faDownload} from '@fortawesome/free-solid-svg-icons'
import { FontAwesomeIcon } from '@fortawesome/vue-fontawesome'

library.add(faSearch,faUpload,faTrash,faFile,faFolder,faDownload,faTrash)

Vue.component('font-awesome-icon', FontAwesomeIcon)
