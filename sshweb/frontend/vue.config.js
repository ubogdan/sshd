
module.exports = {
    outputDir: '../frontendbuild/dist',// build for golang
    devServer: {
        disableHostCheck: true,
        port: 1709,
        https: false,

        //proxy:'http://localhost:2222'
        proxy: {
            '/api/ws/': {
                target: 'ws://127.0.0.1:8022',
                ws: true,
                changeOrigin: true
            },
            '/api': {
                ws: false,
                changeOrigin: true,
                target: 'http://127.0.0.1:8022'
            },
        }
    },
};
