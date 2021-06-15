import axios from 'axios';
import {Message} from 'element-ui';

// 创建axios实例
const service = axios.create({
    //baseURL: 'http://localhost:2222', // api的base_url
    timeout: 120000, // 请求超时时间,
    // request payload
    headers: {
        'Content-Type': 'application/json;charset=UTF-8'
    }
    // 修改请求数据,去除data.q中为空的数据项,只针对post请求
});

service.interceptors.request.use(config => {
    // config.headers['Authorization'] = `Bearer ${store.getters.getJwt}`
    return config;
}, error => {

    return Promise.reject(error);
});


// http response 拦截器
service.interceptors.response.use(res => {
        let {code, data, msg} = res.data;
        if (code === 200) {
            return data
        } else {
            Message.error(msg)
            return false
        }
    },
    error => {
        Message.error(`网络错误`);
        return Promise.reject(error.response.data)
    }
)


export default service


