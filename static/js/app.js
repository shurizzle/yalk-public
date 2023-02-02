import * as Chat from "./chat.js"
import { UserClass, } from "./user.js";
import {
    showModal, messageRow, NewResultHeader,
    WhiteSpacer, noResultRow, elementsHide, UserButton
} from "./elements.js";
import { ready } from "./utils.js";
import { handleEventsQueue } from "./queue.js";
import { profile } from "./settings.js";
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

async function init() {
    const data = { "event": "init" }
    ylk.Queued.push(data)
    let btnChans = document.getElementById("btnChans")
    let btnDms = document.getElementById("btnDms")

    // **** EventListener loader ****
    window.addEventListener("beforeunload", function (e) {
        ylk.Sock.terminate()
    }, false);

    // window.addEventListener('load', function () {
    //     let row_added = false
    //     // !! CHANGE  !!
    //     // $("#user-profile").on('submit', function (e) {
    //     //     e.preventDefault(); // avoid to execute the actual submit of the form.
    //     //     $.ajax({
    //     //         type: $(this).attr('method'),
    //     //         url: $(this).attr('action'),
    //     //         xhrFields: {
    //     //             withCredentials: true
    //     //         },
    //     //         data: $(this).serialize(), // serializes the form's elements.
    //     //         success: function (response) {
    //     //             console.log('Succes: ' + response)
    //     //         },
    //     //         error: function (xhr) {
    //     //             console.log(xhr)
    //     //         }
    //     //     });
    //     // });
    // })

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
        let originID = payload.origin
        let payloadData = payload.data
        let originUser = ylk.Users[originID]
        let originStatusBadge = document.querySelector('#user-profile-' + originID + " > .status-badge")

        if (payload.event === "ping") {
            let calc = Date.now() - payload.data
            document.querySelector("#ping").innerText = calc + "ms"
        }
        if (payload.event === "user_new") {
            let user = payload.data
            ylk.Users[idConn] = UserClass(user)
        }
        if (payload.event === "user_conn") {
            if (ylk.Self == "" || ylk.Self.user_id == originID) {
                return
            }
            ylk.Users[originID].isOnline = true
            originStatusBadge.src = '/static/images/' + originUser.status + '.png'
        }

        if (payload.event === "user_disconn") {
            if (ylk.Self == "" || ylk.Self.user_id == originID) {
                return
            }
            originUser.isOnline = false
            originStatusBadge.src = '/static/images/offline.png'
            return
        }

        if (payload.event === "user_update") {
            let id = payload.origin
            let userData = payload.data
            const selectorId = "#user-profile-" + id

            if (ylk.Users[id] === undefined) {
                return
            }
            ylk.Users[id] = UserClass(userData)
            originStatusBadge.src = '/static/images/' + userData.status + '.png'

            // TODO: Check userUpdate() method on user class
            if (id === ylk.Self.user_id) {
                if (ylk.Context === "PROFILE"){
                    document.querySelector("#profile-avatar").src = "/static/data/user_avatars/" + id + "/avatar.png?" + new Date().getTime()
                }
            document.querySelector("#currentUser > .avatar").src = "/static/data/user_avatars/" + id + "/avatar.png?" + new Date().getTime()
            }
            document.querySelector(selectorId + " > .avatar").src = "/static/data/user_avatars/" + id + "/avatar.png?" + new Date().getTime()

            //TODO: Split in function
            if (id === ylk.Self.user_id) {
                let usrRowStatus = document.querySelector("#btnPopStatus > .status-badge")
                usrRowStatus.src = '/static/images/' + userData.status + '.png'
                elementsHide()
            }
        }

        if (payload.event === "chat_message") {
            let lastMessageID = document.getElementById("receive").lastElementChild.id
            let lastMessage = ylk.Chats[payloadData.to].messages[lastMessageID]
            let sender = ylk.Users[payloadData.from]
            let message_row = messageRow(lastMessage, payloadData.from, payloadData.type, sender.display_name, payloadData.text, sender.admin, sender.color, payloadData.time, payloadData.message_id)
            let receive_area = document.getElementById("receive")

            // Saving message data globally
            ylk.Chats[payloadData.to].messages[payloadData.message_id] = {
                "message_id": payloadData.message_id,
                "from": payloadData.from,
                "to": payloadData.to,
                "type": payloadData.type,
                "text": payloadData.text,
                "time": payloadData.time
            }

            if (ylk.Context != '' && ylk.Context == payloadData.to) {
                receive_area.append(message_row)
                if (payloadData.from != ylk.Self.user_id) {
                    audio.play()
                }
                Chat.Scroll(receive_area)
            }
        }

        if (payload.event === "chat_create") {
            let ctrChans = document.getElementById('sidebar-channels')
            let ctrDms = document.getElementById('sidebar-dms')
            let chat = Chat.New(payloadData, ylk.Users, origin)
            const selfIndex = chat.users.indexOf(ylk.Self.user_id)

            ylk.Chats[chat.id] = chat

            if (selfIndex > -1) {
                ylk.Self.joined_chats.push(chat.id)

                if (chat.type == "dm" && ctrDms.children[chat.id] == undefined) {
                    ctrDms.append(chat.html_button)
                }
                if ((chat.type == "channel_public" || chat.type == "channel_private") && ctrChans.children[chat.id] == undefined) {
                    ctrChans.append(chat.html_button)
                }
            }
        }

        if (payload.event === "chat_delete") {
            let btnChat = document.getElementById(payload.data)

            if (btnChat != undefined) {
                btnChat.remove()
            }

            const index = ylk.Self.joined_chats.indexOf(payload.data)
            if (index > -1) {
                ylk.Self.joined_chats.splice(index, 1)
            }

            delete ylk.Chats[payloadData.data]
            
            Chat.Change(ylk.Chats[ylk.Settings.default_channel])
        }
    }

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

ready(init)