import * as Chat from "./chat.js"
import { UserRow, UserClass, } from "./user.js";
import {
    showModal, MessageRow, NewResultHeader,
    WhiteSpacer, noResultRow, elementsHide, UserButton
} from "./elements.js";
import { DebugBox, DragElement} from "./ngula/debug.js";


// TODO: For ConnStatus use WebSocket readyState?

// (() => {
//     let ylk = {
//         ConnStatus: 0,
//         Queued: [],
//         Settings: {},
//         Users: {},
//         Chats: {},
//         Self: '',
//         Context: '',
//         Sock: '',
//     }
//     let btnChans = document.getElementById("btnChans")
//     let btnDms = document.getElementById("btnDms")

//     waInit()

// })()

window.ylk = {
    ConnStatus: 0,
    Queued: [],
    Settings: {},
    Users: {},
    Chats: {},
    Self: '',
    Context: '',
    Sock: '',
}

window.addEventListener('DOMContentLoaded', function () {
    waInit()
})

async function waInit() {
    let btnChans = document.getElementById("btnChans")
    let btnDms = document.getElementById("btnDms")

    // **** EventListener loader ****
    // * Sidebar newChan button
    btnChans.addEventListener('click', function (e) {
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
    btnDms.addEventListener('click', function (e) {
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

function waEvent() {
    // let websoc_rcv
    var audio = new Audio('static/audio/message.ogg');


    let sb = document.getElementById('sidebar')
    // let ctrChans = document.getElementById('sidebar-channels')
    // let ctrDms = document.getElementById('sidebar-dms')
    let ctrUsers = document.getElementById('server-users')


    // **** Workers launcher ****
    if (typeof (Worker) == "undefined") {
        return false
    }

    // **** Websock worker ****
    ylk.Sock = new Worker("static/js/conn.js")
    ylk.Sock.onmessage = function (event) {
        let payload = event.data

        // * Handling request responses here 

        if (ylk.Queued.length != 0) {
            ylk.Queued.forEach((item, index) => {
                if (item.event == payload.event) {
                    switch (payload.event) {
                        case "chat_create":
                            if (payload.data.type == item.type && payload.origin == ylk.Self.user_id && payload.data.name == item.name) {
                                ylk.Queued.splice(index, 1)
                                ylk.Chats[payload.data.id] = Chat.New(payload.data, ylk.Users, ylk.Self.user_id)
                                Chat.Change(ylk.Chats[payload.data.id])
                                // elementsHide(document.getElementById('modal-box'))
                            }
                            break
                        case "chat_delete":
                            if (payload.data.type == item.type && payload.origin == ylk.Self.user_id && payload.data.name == item.name) {
                                ylk.Queued.splice(index, 1)
                                Chat.Change(ylk.Chats[ylk.Settings.default_channel])
                            }
                            break
                        case "chat_join":
                            // TODO: Check whether chat exists or it's new (channel_private or dm)
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
                            break
                    }
                }
            })
        }
        let msg = payload.data
        switch (payload.event) {
            // !!!! DEBUG !!!! //
            case "ping":
                let calc = Date.now() - payload.data
                document.querySelector("#ping").innerText = calc + "ms"
            // !!!! //
            case "user_conn":
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
                            if (ylk.Self.is_admin == "true") {
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
                document.getElementById('loading-screen').style.display = 'none'
                break
            case "user_update":
                break
            case "chat_message":
                let from = msg.from
                let to = msg.to
                let type = msg.type
                let text = msg.text
                let id = msg.message_id
                let time = msg.time
                let dispName = ylk.Users[from].display_name
                let color = ylk.Users[from].color
                let admin = ylk.Users[from].is_admin

                let lm_id = document.getElementById("receive").lastElementChild.id
                let lm = ylk.Chats[to].messages[lm_id]

                ylk.Chats[to].messages[id] = { from, id, text, time, to, type }

                let row_fragment = MessageRow(lm, from, type, dispName, text, admin, color, time, id)

                var message_row = new DocumentFragment();
                message_row.appendChild(row_fragment)

                let receive_area = document.getElementById("receive")
                if (ylk.Context != '' && ylk.Context == to) {
                    receive_area.append(message_row)
                    if (from != ylk.Self.user_id) {
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

// window.addEventListener("beforeunload", function (e) {
//     ylk.Sock.terminate()
// }, false);

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
})

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