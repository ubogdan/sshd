export const menuUser = [
    {
        txt: '个人中心',
        icon: 'el-icon-c-scale-to-original',
        subs: [
            {path: '/requisition-my', txt: 'Manage User', name: 'requisitionMy'},
        ]
    },
]
export const menuAdmin = [

    {
        txt: 'CMDB',
        icon: 'el-icon-c-scale-to-original',
        subs: [
            {path: '/asset', txt: '资产列表', name: 'asset'},
            {path: '/manage-account', txt: '管理账号', name: 'manageAccount'},
            {path: '/script-exec', txt: '命令执行', name: 'script-exec'},
            {path: '/ssh-session-log', txt: '日志审计', name: '日志审计'},
            {path: '/asset-user', txt: '资产用户', name: '资产用户'},
            // {path: '/manage-account2', txt: '录像审计', name: '录像审计'},//todo:: 录像审计
        ]
    },
    {
        txt: 'Chat',
        icon: 'el-icon-camera',
        subs: [
            {path: '/user', txt: '用户管理', name: 'user'},
        ]
    },
    {
        txt: 'Misc',
        icon: 'el-icon-camera',
        subs: [
            {path: '/hacknews', txt: 'Hack New', name: 'hacknews'},
            // {path: '/mail-center', txt: 'MailCenter', name: 'hacknews'},//todo::
            // {path: '/dev-center', txt: '开发者中心', name: 'DevCenter'},//todo:::
            // {path: '/util/image-editor', txt: '图片编辑', name: 'ImageEditor'},
        ]
    },
]
