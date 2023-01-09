import {
    showModal, newBubbleText, newHashtagText,
    elementsMessageRow, NewDMIcon, NewResultHeader,
    WhiteSpacer, elementsHide, noResultRow
} from "./elements.js";

const origin = window.location.origin
const sse_endpoint = "/events/receive"
const sse_url = origin + sse_endpoint
var audio = new Audio('static/audio/message.ogg');

window.ylk = {
    Settings: {},
    Users: {},
    Chats: {},
    Self: '',
    Context: '',
}

function requestJson(method, url, auth, data) {
    return new Promise(function (res) {
        let xhttp = new XMLHttpRequest()
        if (auth) {
            xhttp.withCredentials = true
        }
        xhttp.open(method, url, true);
        xhttp.setRequestHeader("Content-type", "application/json");
        xhttp.onreadystatechange = function () {
            if (this.readyState == 4 && this.status == 200) {
                res(JSON.parse(this.responseText))
            }
        }
        if (data == undefined) {
            xhttp.send()
        } else {
            xhttp.send(JSON.stringify(data))
        }
    })
}

// **** CHAT ****
class Chat {
    html_button;

    constructor(id, type, name, users, messages) {
        if (messages == null || messages == '') { messages = [] }
        this['id'] = id
        this['type'] = type
        this['name'] = name
        this['users'] = users
        this['messages'] = messages
    }
}

function chatClass(res, srv_usr) {
    let origin = res["origin"]
    let id = res["id"]
    let type = res["type"]
    let name = res["name"]
    let users = res["users"]
    let messages = res["messages"]
    if (res["messages"] != null) { messages = res["messages"] }

    let chat = new Chat(id, type, name, users, messages)

    let button = document.createElement('button')
    button.id = chat.id
    button.className = 'side-item sidebar-item'

    switch (chat.type) {
        case "dm":
            const self_index = chat.users.indexOf(origin)
            let users_display_names = []

            if (self_index > -1) {
                chat.users.splice(self_index, 1)
            }
            chat.users.forEach(element => {
                users_display_names.push(srv_usr[element].display_name)
            })
            var button_items = newBubbleText(users_display_names.toString())
            button.appendChild(button_items.icon_bubble)
            button.appendChild(button_items.bubble_text)
            break

        case "channel_public":
        case "channel_private":
            var button_items = newHashtagText(chat.name)
            button.appendChild(button_items.icon_hashtag)
            button.appendChild(button_items.channel_text)
            break
    }

    button.addEventListener('click', function (e) {
        chatChange(chat)
    })
    chat.html_button = button
    return chat
}




function chatScroll(el) { //? GRAPHIC ELEMENT??
    el.scrollTop = el.scrollHeight
    // if (el.scrollTop != el.scrollHeight() {
    // TODO: This will be needed for detaching live scroll
    // }
}

function chatChange(chat) {
    if (ylk.Context == '') { ylk.Context = chat.id }
    let oldBtn = document.getElementById(ylk.Context)
    let newBtn = document.getElementById(chat.id)

    ylk.Context = chat.id

    document.querySelector("#send").value = ""
    document.querySelector("#receive").innerHTML = ''


    if (oldBtn != undefined) {
        oldBtn.classList.remove('active')
    }
    newBtn.classList.add('active')



    let hdr = document.getElementById("chat-header")
    let hTitle = document.getElementById("header-title")
    hTitle.innerHTML = ''

    let hDelete = document.getElementById("header-delete")
    let new_hDelete = hDelete.cloneNode(true)
    new_hDelete.addEventListener('click', function (e) {
        requestJson("POST", "/chat/delete", true, { "id": chat.id }).then((v) => {
            // TODO: Let the worker take care of this
            let btnChat = document.getElementById(v.data)
            if (btnChat != undefined) {
                btnChat.remove()
            }
            const index = ylk.Self.joined_chats.indexOf(v.data)
            if (index > -1) {
                ylk.Self.joined_chats.splice(index, 1)
            }
            delete ylk.Chats[v.data]
            chatChange(ylk.Chats[ylk.Settings.default_channel])
        })
    })

    hdr.replaceChild(new_hDelete, hDelete)
    let receive_area = document.getElementById("receive")

    switch (chat.type) {
        case "channel_public":
        case "channel_private":
            var ico = newHashtagText(chat.name)
            hTitle.appendChild(ico.icon_hashtag)
            hTitle.appendChild(ico.channel_text)
            break

        case "dm":
            const i = chat.users.indexOf(ylk.Self.user_id)
            if (i > -1) {
                chat.users.splice(i, 1)
            }
            let usrDN = []
            chat.users.forEach(e => {
                usrDN.push(ylk.Users[e].display_name)
            })
            var ico = newBubbleText(usrDN)
            hTitle.appendChild(ico.icon_bubble)
            hTitle.appendChild(ico.bubble_text)
            break
    }
    let chat_messages = []
    if (chat.messages === null || chat.messages === undefined || Object.keys(chat.messages).length === 0) {
        let message = elementsMessageRow("", 0, "server_message", "RosmoBOT", "Empty", true, "", "0")
        chat_messages.push(message)
    } else {
        let lm = null
        for (const [key, value] of Object.entries(chat.messages)) {
            let user_id = value.from
            let display_name = ylk.Users[user_id].display_name
            let is_admin = ylk.Users[user_id].is_admin
            let color = ylk.Users[user_id].color
            let message = elementsMessageRow(lm, user_id, value.type, display_name, value.text, is_admin, color, value.timestamp, key)
            chat_messages.push(message)
            lm = value
        }
    }
    chat_messages.forEach(element => {
        receive_area.appendChild(element)
    })
    // receive_area.scrollIntoView()
    chatScroll(receive_area)
}

async function chatOpen(chType, chName, chUsers, srvData, _chId) {
    let res;
    let reqChat = srvData.Chats[_chId]

    // * If I can find the chat between all those I have fetched (backend only gives ones I am allowed to see)
    if (reqChat != undefined) {
        // * If user data shows it's joined already, just switch to the chat
        if (srvData.Self.joined_chats.includes(reqChat.id)) {
            res = reqChat
        } else {
            // * Else ask the server to join it
            let _res = await requestJson("POST", "/chat/join", true, { "id": _chId })
            res = ylk.Chats[_chId]
            ylk.Self.joined_chats.push(res.id)
            // if (res.data[_chId] != undefined) {
            // TODO: Change to redownload user data 
        }
        // }
    } else {
        let data = {
            "chat_type": chType,
            "users": chUsers,
            "name": chName
        }
        let _res = await requestJson("POST", "/chat/create", true, data)
        _res['data']['origin'] = _res["origin"]
        res = chatClass(_res["data"], ylk.Users)
        ylk.Chats[res.id] = res
        ylk.Self.joined_chats.push(res.id)

    }
    // if (document.getElementById(res.id) == null) { 
    // if (document.querySelector("#"+res.id).parentElement.id == 'search-popup-content') {
    // ? Approach of QuerySelectorAll and checking parentElement.id seems better, less code and can be done with a switch statement.
    let ctrChan = document.getElementById('sidebar-channels')
    let ctrDms = document.getElementById('sidebar-dms')

    if (res.type == "dm" && ctrDms.children[res.id] == undefined) {
        ctrDms.append(res.html_button)
    }
    if ((res.type == "channel_public" || res.type == "channel_private") && ctrChan.children[res.id] == undefined) {
        ctrChan.append(res.html_button)
    }
    chatChange(res)
    elementsHide()
    return res
}

// **** Event  ****
function eventAttach() {
    var btnChans = document.getElementById("btnChans")
    btnChans.addEventListener('click', function (e) {
        let chMdl = showModal("channel");
        document.body.appendChild(chMdl)
        let form = document.querySelector("#modal-form")
        form.addEventListener('submit', function (e) {
            e.preventDefault(); // avoid to execute the actual submit of the form.
            // const formData = new FormData(form)
            const formData = new FormData(form)
            chatOpen("channel_public", formData.get('channel_name'), [], ylk)

            // let modal = document.getElementById("modal-box")
            // if (modal != null) { modal.remove() }
        });
        chMdl.style.display = "flex";
    })
    var btnDms = document.getElementById("btnDms")
    btnDms.addEventListener('click', function (e) {
        let dmMdl = showModal("dms");
        document.body.appendChild(dmMdl)
        let form = document.querySelector("#modal-form")
        form.addEventListener('submit', function (e) {
            e.preventDefault(); // avoid to execute the actual submit of the form.
            const formData = new FormData(form)

            chatOpen("dm", "#", formData.get('users'), ylk)

            // let modal = document.getElementById("modal-box")
            // if (modal != null) { modal.remove() }
        });
        dmMdl.style.display = "flex";
    })

    var searchBox = document.getElementById("searchInput")
    searchBox.addEventListener('keyup', function (e) {
        search(e, ylk.Self)
    })
}




// **** User ****
class User {
    constructor(id, username, display_name, color, status, is_admin, joined_chats) {
        this['user_id'] = id
        this['username'] = username
        this['display_name'] = display_name
        this['color'] = color
        this['user_status'] = status
        this['is_admin'] = is_admin
        this['joined_chats'] = joined_chats
    }
    updateValues() {
        var instance = this
        $.ajax({
            type: 'get',
            url: '/user?' + $.param({ 'id': instance.user_id }),
            xhrFields: {
                withCredentials: true
            },
            success: function (response) {
                var res = JSON.parse(response)
                instance.username = res["username"]
                instance.display_name = res["display_name"]
                instance.color = res["color"]
                instance.status = res["status"]
                instance.is_admin = res["is_admin"]
                instance.joined_chats = res["joined_chats"]
                var user_row_id = 'user-profile-' + instance.user_id
                var badge_selector = user_row_id + ' > img.status-badge'
                $('#' + badge_selector).attr('src', 'static/images/' + instance.status + '.png')
            },
            error: function (xhr) {
                console.log(xhr)
            }
        })
    }
}

function userClass(res) {

    let user_id = res["id"]
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

function userRow(current_user, userData) {
    var user_row = document.createElement('div')
    var user_avatar = document.createElement('img')
    user_avatar.className = 'avatar'
    user_avatar.src = 'static/data/user_avatars/' + userData.user_id + '/avatar.png'

    user_avatar.style.borderColor = userData.color


    var status_dot = document.createElement('img')
    status_dot.className = 'status-badge'
    status_dot.src = 'static/images/' + userData.user_status + '.png'

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

function userButton(userData, input, sel_ctr) {
    var user_row_btn = document.createElement('button')
    var user_avatar = document.createElement('img')
    user_avatar.className = 'avatar'
    user_avatar.src = 'static/data/user_avatars/' + userData.user_id + '/avatar.png'

    user_avatar.style.borderColor = userData.color

    var status_dot = document.createElement('img')
    status_dot.className = 'status-badge'
    status_dot.src = 'static/images/' + userData.user_status + '.png'

    var username = document.createElement('span')
    username.className = 'username'
    username.innerText = userData.display_name

    if ((input != undefined && input != null) && (sel_ctr != undefined && sel_ctr != null)) {
        user_row_btn.id = 'user-profile-' + userData.user_id
        user_row_btn.classList = 'profile-row btn-fw btn-inv'

    } else {
        user_row_btn.id = userData.user_id
        user_row_btn.classList = 'profile-row btn-fw btn-inv'
        user_row_btn.addEventListener("click", function (input) {
            input.value = ""
            $("#search-popup").get(0).style.display = 'none'
            let user = document.createElement('span')
            user.innerText = userData.user_id
        })
    }
    user_row_btn.appendChild(user_avatar)
    user_row_btn.appendChild(status_dot)
    user_row_btn.appendChild(username)
    var dmIcon = NewDMIcon(userData.user_id)
    user_row_btn.appendChild(dmIcon)

    return user_row_btn
}



// **** WebApp (wa) ****
async function waLoad() {
    let sb = document.getElementById('sidebar')
    let chanCtr = document.getElementById('sidebar-channels')
    let dmCtr = document.getElementById('sidebar-dms')
    let usrCtr = document.getElementById('server-users')

    let _usr = await requestJson("GET", "/user/all", true)
    _usr['data'].forEach(element => {
        let _cu = userClass(element)
        ylk.Users[_cu.user_id] = _cu
    })

    let _self = await requestJson("GET", "/user", true)
    ylk.Self = userClass(_self['data'])

    let _set = await requestJson("GET", "/settings", true)
    for (const [key, value] of Object.entries(_set['data'])) {
        ylk.Settings[key] = value
    }

    for (const [_, v] of Object.entries(ylk.Users)) {
        let _usr = userRow(false, v)
        usrCtr.appendChild(_usr)
    }

    let _chat = await requestJson("GET", "/chat/all", true)
    for (const [k, v] of Object.entries(_chat['data'])) {
        v['origin'] = _chat["origin"]
        ylk.Chats[k] = chatClass(v, ylk.Users)
    }

    ylk.Self.joined_chats.forEach(element => {
        let chat = ylk.Chats[element]
        switch (chat.type) {
            case "channel_public":
            case "channel_private":
                chanCtr.append(chat.html_button)
                break
            case "dm":
                dmCtr.append(chat.html_button)
                break
        }
    })
    // * Own profile tab loader
    sb.appendChild(userRow(true, ylk.Self))
    chatChange(ylk.Chats[ylk.Settings.default_channel])
    document.getElementById('loading-screen').style.display = 'none'
}

function waConnect() {
    if (typeof (Worker) == "undefined") {
        return false
    }
    let sse_rcv = new Worker("static/js/w/receiver.js")
    sse_rcv.postMessage({ "args": [sse_url] })
    sse_rcv.onmessage = function (event) {
        eventLog(event)
    };
    window.addEventListener("beforeunload", function (e) {
        sse_rcv.terminate()
    }, false);

    let websoc_rcv = new Worker("static/js/w/websocket.js")
    // websoc_rcv.postMessage({ "args": [sse_url] })
    websoc_rcv.onmessage = function (event) {
        eventLog(event)
    };
    window.addEventListener("beforeunload", function (e) {
        websoc_rcv.terminate()
    }, false);
}

function eventLog(res) {
    let msg = res.data
    if (msg.success != true) {
        return "err"
    }
    switch (msg.event) {
        case "status_update":
            break
        case "chat_message":


            let new_mess = eventLog(msg)


            let from = msg.from
            let to = msg.to
            let type_message = msg.type
            let text = msg.text
            var message_id = msg.id
            let timestamp = msg.timestamp
            let display_name = ylk.Users[from].display_name
            let color = ylk.Users[from].color
            let is_admin = ylk.Users[from].is_admin

            let lm_id = document.getElementById("receive").lastElementChild.id
            var lm = ylk.Chats[to].messages[lm_id]

            ylk.Chats[to].messages[message_id] = { from, message_id, text, timestamp, to, type_message }

            let row_fragment = elementsMessageRow(lm, from, type_message, display_name, text, is_admin, color, timestamp, message_id)
            // let rcv_event = msg.data.to

            var message_row = new DocumentFragment();
            message_row.appendChild(row_fragment)

            let origin = res.data.origin
            let receive_area = document.getElementById("receive")
            if (ylk.Context != '' && ylk.Context == to) {
                receive_area.append(new_mess)
                if (origin != ylk.Self.user_id) {
                    audio.play()
                }
                // document.getElementById("receive").lastElementChild.scrollIntoView()
                chatScroll(receive_area)
            }
            break
        case "chat_create":
            msg['data']['origin'] = msg["origin"]
            let res = chatClass(msg["data"], ylk.Users)
            ylk.Chats[res.id] = res
            // ! If I'm between the users add between my joined
            ylk.Self.joined_chats.push(res.id)
            // !!!!!!
            break
        case "chat_delete":
            let btnChat = document.getElementById(msg.data)
            if (btnChat != undefined) {
                btnChat.remove()
            }
            const index = ylk.Self.joined_chats.indexOf(msg.data)
            if (index > -1) {
                ylk.Self.joined_chats.splice(index, 1)
            }
            delete ylk.Chats[msg.data]
            chatChange(ylk.Chats[ylk.Settings.default_channel])
            break
        default:
            break
    }
}

window.addEventListener('DOMContentLoaded', function () {
    waLoad()
})

window.addEventListener('load', function () {
    let row_added = false
    // !! CHANGE  !!
    $("#user-profile").on('submit', function (e) {
        e.preventDefault(); // avoid to execute the actual submit of the form.
        $.ajax({
            type: $(this).attr('method'),
            url: $(this).attr('action'),
            xhrFields: {
                withCredentials: true
            },
            data: $(this).serialize(), // serializes the form's elements.
            success: function (response) {
                console.log('Succes: ' + response)
            },
            error: function (xhr) {
                console.log(xhr)
            }
        });
    });
    // !! CHANGE  !!
    const send = document.getElementById("send")
    send.addEventListener("keydown", async function (key) {
        if ((key.code === "Enter" || key.code === "NumpadEnter")) {
            key.preventDefault()
            let msg_text = send.value
            let data = {
                "chat_id": ylk.Context,
                "message": msg_text,
            }
            if (ylk.Settings.conn_type == 'ws') {
                try {
                    const resp = await fetch("/websocket/send", {
                        method: "POST",
                        body: {
                            "chat_id": ylk.Context,
                            "message": msg_text,
                        },
                    })
                    if (resp.status !== 202) {
                        throw new Error("Unexpected HTTP Status " + resp.status + " " + resp.statusText)
                    }
                } catch (err) {
                    console.log("Publish failed: " + err.message)
                }
            } else if (ylk.Settings.conn_type == 'sse') {
                send.value = ""
                $.ajax({
                    url: "/events/send",
                    type: "post",
                    xhrFields: {
                        withCredentials: true
                    },
                    data: {
                        chat_id: ylk.Context,
                        message: msg_text,
                    },
                    success: function (response) {

                    },
                    error: function (xhr) {
                        console.log(xhr)
                    }
                });
            }
        }
    }, false);
    eventAttach()
    waConnect()
})

function search(element) {
    var input, kp, filter, txtValue, searchPopup;
    input = element.target
    kp = element.key
    searchPopup = document.getElementById("search-popup");
    let searchPopupContent = document.getElementById("search-popup-content");
    searchPopupContent.innerHTML = ''
    if (input.value == "") {
        searchPopup.style.display = "none"
    } else {
        searchPopup.style.display = "block"
        filter = input.value.toUpperCase();

        // ******* DMs search box ******* //
        var usrCtr = NewResultHeader("Users")
        var usr_found = false;

        for (const [key, value] of Object.entries(ylk.Users)) {
            let txtValue = value.display_name
            // txtValue.toUpperCase().indexOf(filter)

            if (txtValue.toUpperCase().indexOf(filter) > -1) {
                usr_found = true
                let user_row_btn = userButton(value)
                user_row_btn.addEventListener("click", function (e) {
                    chatOpen("dm", "#", key, ylk)
                })
                usrCtr.appendChild(user_row_btn)
            }
        }
        if (!usr_found) {
            usrCtr.appendChild(noResultRow())
        }
        // ******* Channels search box ******* //
        var chanCtr = NewResultHeader("Channels")
        var chat_found = false;

        for (const [key, value] of Object.entries(ylk.Chats)) {
            let txtValue = value.name
            txtValue.toUpperCase().indexOf(filter)

            if (txtValue.toUpperCase().indexOf(filter) > -1) {
                chat_found = true
                let search_chan_btn = value.html_button.cloneNode(true)
                search_chan_btn.addEventListener('click', function (e) {
                    chatOpen("channel_public", value.name, value.users, ylk, value.id)
                })
                chanCtr.appendChild(search_chan_btn)
            }
        }
        if (!chat_found) {
            var no_result = noResultRow()
            chanCtr.appendChild(no_result)
        }
        searchPopupContent.appendChild(usrCtr)
        searchPopupContent.appendChild(WhiteSpacer())
        searchPopupContent.appendChild(chanCtr)
    }
}