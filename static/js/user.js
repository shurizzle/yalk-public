import { NewDMIcon, StatusPopup } from "./elements.js"
import * as settings from "./settings.js"
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
    // get isOnline() {return this.isOnline}
    // /**
    //  * @param {bool} s Set user description
    //  */
    // set isOnline(s) {
    //     console.info("Socio sto settando")
    //     this.isOnline = s
    //     let statusBadge = document.querySelector('#user-profile-' + this.user_id + " > .status-badge")
    //         statusBadge.src = 'static/images/' + this.status + '.png'
    // }
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
// customElements.define('user-profile', settings.Profile)

export function UserRow(isSelf, userData) {
    var row = document.createElement('div')
    var avatar = document.createElement('img')
    avatar.addEventListener('click', () => {
        ylk.Context = "PROFILE"
        const hTitle = document.getElementById("header-title")
        let hDelete = document.getElementById("header-delete")
        let receiveArea = document.getElementById("receive")
        let oldBtn = document.getElementById(ylk.Context)

        receiveArea.innerHTML = ''
        document.querySelector("#send").value = ''
        hTitle.innerHTML = 'Settings'
        hDelete.style.visibility = "hidden"

        // oldBtn.style.visibility = 'hidden' // ! remove active class
        document.querySelector("#send").style.visibility = "hidden"
        document.querySelector(".chat-grid-divider").style.visibility = "hidden"


        let profile = settings.profile(userData)

        // profile.querySelector('#picture-submit').addEventListener('click', )
        receiveArea.appendChild(profile)

    })
    avatar.className = 'avatar'
    avatar.src = 'static/data/user_avatars/' + userData.user_id + '/avatar.png'

    avatar.style.borderColor = userData.color


    var statusDot = document.createElement('img')
    statusDot.className = 'status-badge'
    if (userData.isOnline == false) {
        statusDot.src = 'static/images/offline.png'
    } else {
        statusDot.src = 'static/images/' + userData.status + '.png'
    }

    var username = document.createElement('span')
    username.className = 'username'
    username.innerText = userData.display_name

    if (isSelf) {
        row.id = 'currentUser'
        row.classList = 'profile-row'

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

        status_btn.appendChild(statusDot)


        row.appendChild(avatar)
        row.appendChild(status_btn)
        row.appendChild(username)
        row.appendChild(logout_link)

    }
    else {
        row.id = 'user-profile-' + userData.user_id
        row.classList = 'profile-row'
        row.appendChild(avatar)
        row.appendChild(statusDot)
        row.appendChild(username)
        var dmIcon = NewDMIcon(userData.user_id)
        row.appendChild(dmIcon)
    }

    return row
}

