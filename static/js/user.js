import { NewDMIcon } from "./elements.js"

// **** User ****
class User {
    constructor(id, username, display_name, color, status, is_admin, joined_chats) {
        this['user_id'] = id
        this['username'] = username
        this['display_name'] = display_name
        this['color'] = color
        this['status'] = status
        this['is_admin'] = is_admin
        this['joined_chats'] = joined_chats
    }
    // get status() {return this.status}
    // set status(s) {this.status = s}
    // updateValues() {
    //     var instance = this
    //     $.ajax({
    //         type: 'get',
    //         url: '/user?' + $.param({ 'id': instance.user_id }),
    //         xhrFields: {
    //             withCredentials: true
    //         },
    //         success: function (response) {
    //             var res = JSON.parse(response)
    //             instance.username = res["username"]
    //             instance.display_name = res["display_name"]
    //             instance.color = res["color"]
    //             instance.status = res["status"]
    //             instance.is_admin = res["is_admin"]
    //             instance.joined_chats = res["joined_chats"]
    //             var user_row_id = 'user-profile-' + instance.user_id
    //             var badge_selector = user_row_id + ' > img.status-badge'
    //             $('#' + badge_selector).attr('src', 'static/images/' + instance.status + '.png')
    //         },
    //         error: function (xhr) {
    //             console.log(xhr)
    //         }
    //     })
    // }
    
}

export function UserClass(res) {
    let user_id = res["user_id"]
    let username = res["username"]
    let display_name = res["display_name"]
    let color = res["color"]
    let status = res["status"]
    let is_admin = res["is_admin"]
    let joined_chats = res["joined_chats"]
    if (res["joined_chats"] != null) { joined_chats = res["joined_chats"] }
    return new User(user_id, username, display_name, color, status, is_admin, joined_chats)
}

function userStatusUpdate(status) {
    $.ajax({
        url: "/user/update/status",
        type: "post",
        xhrFields: {
            withCredentials: true
        },
        data: {
            event: "status_update",
            status: status,
        },
        success: function (response) {
            ylk.Self.status
            $('.btn-open img.status-badge').attr('src', 'static/images/' + status + '.png')
            document.getElementById("user-status-popup").style.display = "none";
        },
        error: function (xhr) {
            console.log(xhr)
        }
    });
}

export function UserRow(current_user, userData) {
    var user_row = document.createElement('div')
    var user_avatar = document.createElement('img')
    user_avatar.addEventListener('click', () => {
        window.location.replace('/profile')
    })
    user_avatar.className = 'avatar'
    user_avatar.src = 'static/data/user_avatars/' + userData.user_id + '/avatar.png'

    user_avatar.style.borderColor = userData.color


    var status_dot = document.createElement('img')
    status_dot.className = 'status-badge'
    status_dot.src = 'static/images/' + userData.status + '.png'

    var username = document.createElement('span')
    username.className = 'username'
    username.innerText = userData.display_name

    if (current_user) {
        user_row.id = 'currentUser'
        user_row.classList = 'profile-row'

        // var avatar_link = document.createElement('a')
        // avatar_link.href = '/user'

        var status_btn = document.createElement('button')
        status_btn.id = 'btnPopStatus'
        status_btn.classList = 'btn-open btn-status'
        status_btn.addEventListener('click', function (e) {
            showStatusPopup();
        })

        var logout_link = document.createElement('a')
        logout_link.href = '/logout'
        var logout_logo = document.createElement('i')
        logout_logo.classList = 'fa-solid fa-right-from-bracket'
        logout_link.append(logout_logo)

        status_btn.appendChild(status_dot)


        user_row.appendChild(user_avatar)
        user_row.appendChild(status_btn)
        user_row.appendChild(username)
        user_row.appendChild(logout_link)

    }
    else {
        user_row.id = 'user-profile-' + userData.user_id
        user_row.classList = 'profile-row'
        user_row.appendChild(user_avatar)
        user_row.appendChild(status_dot)
        user_row.appendChild(username)
        var dmIcon = NewDMIcon(userData.user_id)
        user_row.appendChild(dmIcon)
    }

    return user_row
}

