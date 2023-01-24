import { NewDMIcon, StatusPopup } from "./elements.js"

// **** User ****
class User {
    constructor(id, username, display_name, color, status, statusText, isAdmin, joined_chats, isOnline) {
        this['user_id'] = id
        this['username'] = username
        this['display_name'] = display_name
        this['color'] = color
        this['status'] = status
        this['statusText'] = statusText
        this['isAdmin'] = isAdmin
        this['joined_chats'] = joined_chats
        this['isOnline'] = isOnline
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
    //             instance.isAdmin = res["isAdmin"]
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
    let userId = res["user_id"]
    let username = res["username"]
    let displayName = res["display_name"]
    let color = res["color"]
    let status = res["status"]
    let statusText = res["statusText"]
    let isAdmin = res["isAdmin"]
    let joinedChats = res["joined_chats"]
    let isOnline = res["isOnline"]
    if (res["joined_chats"] != null) { joinedChats = res["joined_chats"] }
    return new User(userId, username, displayName, color, status, statusText, isAdmin, joinedChats, isOnline)
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
    if (userData.isOnline == false) {
        status_dot.src = 'static/images/offline.png'
    } else {
        status_dot.src = 'static/images/' + userData.status + '.png'
    }

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
            let statusPopup = StatusPopup();
            document.body.appendChild(statusPopup)
            // document.getElementById("user-status-popup").style.display = "block";
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

