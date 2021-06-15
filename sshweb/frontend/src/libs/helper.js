import store from "@/libs/store";


export function hasApprovePermission(requisitionObj) {
    return (hasApprovePermissionBy(requisitionObj, 1) && requisitionObj.approve_status_1 === '') || (hasApprovePermissionBy(requisitionObj, 2) && requisitionObj.approve_status_1 === 'agreed')
}


export function hasApprovePermissionBy(requisitionObj, lvl) {
    let flag = false
    let user = store.getters.getUser
    if (!user) {
        return flag
    }
    if (!requisitionObj.approve_flow) {
        return flag
    }
    let level = `approves_${lvl}`
    let users = requisitionObj.approve_flow[level]
    if (!users) {
        return flag
    }
    for (const item of users) {
        if (item.id === user.id) {
            flag = true
        }
    }
    return flag
}

//前台将RFC3339时间格式转换为正常格式
export function dateToString(timeString) {
    let date = new Date(timeString).toJSON();
    let timeZone = 8 * 3600 * 1000
    return new Date(+new Date(date) + timeZone).toISOString().replace(/T/g, ' ').replace(/\.[\d]{3}Z/, '');
}

export function timeFormat(s) {
    return s.substr(2, 16).replace('T', ' ')
}