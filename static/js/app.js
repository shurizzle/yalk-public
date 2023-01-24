import * as Chat from "./chat.js"
import { UserRow, UserClass, } from "./user.js";
import {
    showModal, MessageRow, NewResultHeader,
    WhiteSpacer, noResultRow, elementsHide, UserButton
} from "./elements.js";
import { DebugBox, DragElement } from "./ngula/debug.js";


// TODO: For ConnStatus use WebSocket readyState?

// (() => {
//     let ylk = {


//     waInit()
// })()

window.ylk = {
    Init: false,
    Queued: [],
    Settings: {},
    Users: {},
    Chats: {},
    Self: '',
    Context: '',
    Sock: '',
}


window.addEventListener('DOMContentLoaded', function () {
    init()
})

function handleEventsQueue(payload) {
    // if (ylk.Queued.length == 0) {
    //     return
    // }
    let sb = document.getElementById('sidebar')
    let ctrUsers = document.getElementById('server-users')

    ylk.Queued.forEach((item, index) => {
        if (item.event != payload.event) {
            return
        }
        let msg = payload.data
        if (payload.event == "init") {
            for (const [k, v] of Object.entries(msg)) {
                switch (k) {
                    case "users":
                        for (const [key, value] of Object.entries(v)) {
                            let _cu = UserClass(value)
                            ylk.Users[key] = value
                            let _usrRow = UserRow(false, _cu)
                            ctrUsers.appendChild(_usrRow)
                        }
                        break
                    case "self":
                        ylk.Self = UserClass(v)
                        sb.appendChild(UserRow(true, v))
                        if (ylk.Self.isAdmin == "true") {
                            let debugBox = DebugBox()
                            document.body.prepend(debugBox)
                            // document.body.appendChild(debugBox)
                            DragElement(debugBox)
                        }
                        break
                    case "settings":
                        for (const [key, value] of Object.entries(v)) {
                            ylk.Settings[key] = value
                        }
                        break
                    case "chats":
                        for (const [key, value] of Object.entries(v)) {
                            ylk.Chats[key] = Chat.New(value, msg["users"], msg["self"].user_id)
                        }
                        break
                }
            }
            ylk.Self.joined_chats.forEach(element => {
                let chat = ylk.Chats[element]
                switch (chat.type) {
                    case "channel_public":
                    case "channel_private":
                        document.querySelector('#sidebar-channels').append(chat.html_button)
                        break
                    case "dm":
                        document.querySelector('#sidebar-dms').append(chat.html_button)
                        break
                }
            })
            Chat.Change(ylk.Chats[ylk.Settings.default_channel])
            document.querySelector('#loading-screen').style.display = 'none'
            return
        }

        if (payload.event == "chat_create") {
            if (payload.data.type != item.type && payload.origin != ylk.Self.user_id && payload.data.name != item.name) {
                return
            }
            ylk.Queued.splice(index, 1)
            ylk.Chats[payload.data.id] = Chat.New(payload.data, ylk.Users, ylk.Self.user_id)
            Chat.Change(ylk.Chats[payload.data.id])
            return
        }
        if (payload.event == "chat_delete") {
            if (payload.data.type != item.type && payload.origin != ylk.Self.user_id && payload.data.name != item.name) {
                return
            }
            ylk.Queued.splice(index, 1)
            Chat.Change(ylk.Chats[ylk.Settings.default_channel])
            return
        }
        if (payload.event == "chat_join") {
            let chat = ylk.Chats[payload.data.id]
            if (payload.data.id == item.id) {
                ylk.Self.joined_chats.push(payload.data.id)
            }
            if (payload.data.type == "channel_public" || payload.data.type == "channel_private") {
                document.querySelector('#sidebar-channel}}s').append(chat.html_button)
            } else {
                document.querySelector('#sidebar-dms').append(chat.html_button)
            }
            Chat.Change(chat)
        }
    })
}

async function init() {
    const data = {"event": "init"}
    ylk.Queued.push(data)
    let btnChans = document.getElementById("btnChans")
    let btnDms = document.getElementById("btnDms")

    // **** EventListener loader ****

    window.addEventListener("beforeunload", function (e) {
        ylk.Sock.terminate()
    }, false);

    window.addEventListener('load', function () {
        let row_added = false
        // !! CHANGE  !!
        // $("#user-profile").on('submit', function (e) {
        //     e.preventDefault(); // avoid to execute the actual submit of the form.
        //     $.ajax({
        //         type: $(this).attr('method'),
        //         url: $(this).attr('action'),
        //         xhrFields: {
        //             withCredentials: true
        //         },
        //         data: $(this).serialize(), // serializes the form's elements.
        //         success: function (response) {
        //             console.log('Succes: ' + response)
        //         },
        //         error: function (xhr) {
        //             console.log(xhr)
        //         }
        //     });
        // });
    })

    // * Sidebar newChan button
    btnChans.addEventListener('click', () => {
        let chMdl = showModal("channel");
        document.body.appendChild(chMdl)
        // let form = document.querySelector("#modal-form")
        // form.addEventListener('submit', function (e) {
        //     e.preventDefault();
        //     const formData = new FormData(form)
        //     let data = Chat.Open("channel_public", formData.get('channel_name'), [], ylk)
        //     ylk.Queued.push(data)
        //     elementsHide()
        // });
        chMdl.style.display = "flex";
    })

    // * Sidebar newDm button
    btnDms.addEventListener('click', () => {
        let dmMdl = showModal("dms");
        document.body.appendChild(dmMdl)
        let form = document.querySelector("#modal-form")
        form.addEventListener('submit', (ev) => {
            ev.preventDefault(); // avoid to execute the actual submit of the form.
            const formData = new FormData(form)
            let data = Chat.Open("dm", "#", formData.get('users'), ylk)
            if (data === null) {
                elementsHide(ev.target)
                return
            }
            Chat.Change(data)
        });
        dmMdl.style.display = "flex";
    })

    // * Sidebar searchBar button
    let searchBox = document.getElementById("searchInput")
    searchBox.addEventListener('keyup', function (e) {
        search(e, ylk.Self)
    })
    waEvent()

}

// * Handles event received by verifying succesful return of user actions
// * and processing all other events
function waEvent() {
    let audio = new Audio('static/audio/message.ogg');


    // **** Workers launcher ****
    if (typeof (Worker) == "undefined") {
        return false
    }

    // **** Websock worker ****
    ylk.Sock = new Worker("static/js/conn.js")
    ylk.Sock.onmessage = (event) => {
        let payload = event.data

        // * Checking user sent events in 
        // TODO: Expect return
        handleEventsQueue(payload)

        // * Handling received payload
        let msg = payload.data
        switch (payload.event) {
            // !!!! DEBUG !!!! //
            case "ping":
                let calc = Date.now() - payload.data
                document.querySelector("#ping").innerText = calc + "ms"

            case "user_conn":
                let idConn = payload.origin
                let userConn = payload.data
                if (ylk.Users[idConn] == undefined && ylk.Self != "") {
                    ylk.Users[idConn] = UserClass(payload.data)
                    let statusBadge = document.querySelector('#user-profile-' + idConn + " > .status-badge")
                    statusBadge.src = 'static/images/' + userConn.status + '.png'
                }

            case "user_disconn":
                let idDisconn = payload.origin
                if (ylk.Users[idDisconn] !== undefined && ylk.Self.user_id != idDisconn) {
                    ylk.Users[idDisconn].status = "offline"
                    let statusBadge = document.querySelector('#user-profile-' + idDisconn + " > .status-badge")
                    statusBadge.src = 'static/images/offline.png'
                }
                break
            case "user_update":
                let idUpd = payload.origin
                let userUpd = payload.data
                if (ylk.Users[idUpd] === undefined) {
                    break
                }
                ylk.Users[idUpd] = UserClass(userUpd)
                let statusBadge = document.querySelector('#user-profile-' + idUpd + " > .status-badge")
                statusBadge.src = 'static/images/' + userUpd.status + '.png'

                if (idUpd === ylk.Self.user_id) {
                    let usrRowStatus = document.querySelector("#btnPopStatus > .status-badge")
                    usrRowStatus.src = 'static/images/' + userUpd.status + '.png'
                    elementsHide()
                }
                break

            case "chat_message":
                let lastMessageID = document.getElementById("receive").lastElementChild.id
                let lastMessage = ylk.Chats[msg.to].messages[lastMessageID]
                let sender = ylk.Users[msg.from]

                let message_row = MessageRow(lastMessage, msg.from, msg.type, sender.display_name, msg.text, sender.admin, sender.color, msg.time, msg.message_id)

                // Saving message data globally
                ylk.Chats[msg.to].messages[msg.message_id] = {
                    "message_id": msg.message_id,
                    "from": msg.from,
                    "to": msg.to,
                    "type": msg.type,
                    "text": msg.text,
                    "time": msg.time
                }

                let receive_area = document.getElementById("receive")
                if (ylk.Context != '' && ylk.Context == msg.to) {
                    receive_area.append(message_row)
                    if (msg.from != ylk.Self.user_id) {
                        audio.play()
                    }
                    Chat.Scroll(receive_area)
                }
                break

            case "chat_create":
                let ctrChans = document.getElementById('sidebar-channels')
                let ctrDms = document.getElementById('sidebar-dms')
                let chatData = payload.data
                let res = Chat.New(chatData, ylk.Users, origin)
                ylk.Chats[res.id] = res

                const selfIndex = res.users.indexOf(ylk.Self.user_id)
                if (selfIndex > -1) {
                    ylk.Self.joined_chats.push(res.id)
                    if (res.type == "dm" && ctrDms.children[res.id] == undefined) {
                        ctrDms.append(res.html_button)
                    }
                    if ((res.type == "channel_public" || res.type == "channel_private") && ctrChans.children[res.id] == undefined) {
                        ctrChans.append(res.html_button)
                    }
                }
                break

            case "chat_delete":
                let btnChat = document.getElementById(payload.data)
                if (btnChat != undefined) {
                    btnChat.remove()
                }
                const index = ylk.Self.joined_chats.indexOf(payload.data)
                if (index > -1) {
                    ylk.Self.joined_chats.splice(index, 1)
                }
                delete ylk.Chats[msg.data]
                Chat.Change(ylk.Chats[ylk.Settings.default_channel])
                break
            default:
                break
        }

    };

    const send = document.getElementById("send")
    send.addEventListener("keydown", async function (key) {
        if ((key.code === "Enter" || key.code === "NumpadEnter")) {
            key.preventDefault()
            // if (send.value == '') {
            //     return
            // }
            let msg_text = send.value
            let data = {
                "event": "chat_message",
                "to": ylk.Context,
                "text": msg_text,
            }
            ylk.Sock.postMessage(data)
            send.value = ''
        }
    })
}

function search(element) {
    let input = element.target
    let filter = input.value.toUpperCase()
    let keyPress = element.key
    let searchPopup = document.getElementById("search-popup")
    let searchPopupContent = document.getElementById("search-popup-content")
    searchPopupContent.innerHTML = ''
    let usr_found = false
    let chat_found = false

    if (input.value == "") {
        searchPopup.style.display = "none"
        return
    }

    let ctrUsers = NewResultHeader("Users")
    let ctrChans = NewResultHeader("Channels")

    searchPopup.style.display = "block"

    // ******* DMs search box ******* //
    for (const [key, value] of Object.entries(ylk.Users)) {
        if (value.display_name.toUpperCase().indexOf(filter) > -1) {
            usr_found = true
            let user_row_btn = UserButton(value, searchPopup)
            user_row_btn.addEventListener("click", function (e) {
                let data = Chat.Open("dm", "#", key, ylk)
                if (data === null) {
                    elementsHide(searchPopup)
                    return
                }
                Chat.Change(data)
            })
            ctrUsers.appendChild(user_row_btn)
        }
    }

    // ******* Channels search box ******* //
    for (const [_, value] of Object.entries(ylk.Chats)) {
        if (value.name.toUpperCase().indexOf(filter) > -1) {
            chat_found = true
            let search_chan_btn = value.html_button.cloneNode(true)
            search_chan_btn.addEventListener('click', () => {
                let chat = Chat.Open("channel_public", value.name, value.users, ylk, value.id)
                if (chat === null) {
                    elementsHide(ev.target)
                    return
                }
                Chat.Change(chat)
            })
            ctrChans.appendChild(search_chan_btn)
        }
    }
    if (!usr_found) { ctrUsers.appendChild(noResultRow()) }
    if (!chat_found) { ctrChans.appendChild(noResultRow()) }

    searchPopupContent.appendChild(ctrUsers)
    searchPopupContent.appendChild(WhiteSpacer())
    searchPopupContent.appendChild(ctrChans)
}

function UserStatusUpdate(status) {
    let data = {
        "event": "user_update",
        "status": status
    }
    ylk.Self.status = status

    ylk.Queued.push(data)

}