import * as Chat from "./chat.js"

export function showModal(modalType) { // Change
    var modal = document.createElement('div')
    modal.className = 'modal'
    modal.id = 'modal-box'

    // * Button close //
    var btnClose = document.createElement('span')
    btnClose.className = 'modal-close'
    btnClose.addEventListener('click', function (e) {
        modal.style.display = "none";
        modal_content.innerHTML = ''
    })
    var close_ico = document.createElement('i')
    close_ico.classList = 'fa-solid fa-circle-xmark'
    btnClose.appendChild(close_ico)

    // * Modal Content //
    var modal_content = document.createElement("div")
    modal_content.className = 'modal-content'
    modal_content.appendChild(btnClose)

    let form = document.createElement("form")
    form.action = "/chat/create"
    form.method = "POST"
    form.id = "modal-form"
    form.className = 'flex-column'
    form.addEventListener('submit', (ev) => {
        ev.preventDefault();
        const formData = new FormData(form)
        let data = Chat.Open("channel_public", formData.get('channel_name'), [], ylk)
        if (data === null) {
            elementsHide(modal)
            return
        }
        Chat.Change(data)
    });


    let title = document.createElement("label")
    let chatType, btnSubmit;

    switch (modalType) {
        case "channel":
            chatType = "channel_public"
            title.innerText = "New Channel"

            var lbl_name = document.createElement("label")
            lbl_name.innerText = "Channel Name"
            var name = document.createElement("input")
            name.type = 'text'
            name.name = 'channel_name'
            form.appendChild(lbl_name)
            form.appendChild(name)
            btnSubmit = NewSubmitButton("Create")
            form.appendChild(btnSubmit)
            break

        case "dms":
            chatType = "dm"
            title.innerText = "New Direct Message"

            var lbl_sel = document.createElement("label")
            lbl_sel.innerText = "Selected"

            var sel_ctr = document.createElement("div")
            sel_ctr.id = "sel_ctr"


            var lbl_name = document.createElement("label")
            lbl_name.innerText = "Username"

            var name = document.createElement("input")
            btnSubmit = NewSubmitButton("Create")

            name.type = 'text'
            name.name = 'dms_name'
            name.addEventListener('keyup', function (e) {
                UserButton(e, sel_ctr)
            })

            form.appendChild(title)
            form.appendChild(lbl_sel)
            form.appendChild(sel_ctr)
            form.appendChild(lbl_name)
            form.appendChild(name)
            form.appendChild(btnSubmit)
            break


    }
    modal_content.append(form)
    modal.append(modal_content)

    return modal

}

function createModalUpdateUser(id, display_name) {
    var btnClose = document.createElement('span')
    btnClose.className = 'modal-close'
    btnClose.addEventListener('click', function (e) {
        modal.style.display = "none";
        modal_content.innerHTML = ''
    })
    var close_ico = document.createElement('i')
    close_ico.classList = 'fa-solid fa-circle-xmark'
    btnClose.appendChild(close_ico)

    var m_form = document.createElement("form")
    m_form.id = "modal-user-edit"
    m_form.action = "/admin/user/update"
    m_form.method = "POST"
    m_form.className = 'flex-column'

    m_form.addEventListener('submit', function (e) {
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

    var user_id_l = document.createElement("label")
    user_id_l.innerText = "User ID"

    var m_user_id = document.createElement("input")
    m_user_id.type = 'text'
    m_user_id.name = 'user_id'
    m_user_id.readOnly = true
    m_form.appendChild(user_id_l)
    m_form.appendChild(m_user_id)

    var username_l = document.createElement("label")
    username_l.innerText = "Username"
    var m_username = document.createElement("input")
    m_username.type = 'text'
    m_username.name = 'username'
    m_form.appendChild(username_l)
    m_form.appendChild(m_username)

    var display_name_l = document.createElement("label")
    display_name_l.innerText = "Display name"
    var m_display_name = document.createElement("input")
    m_display_name.type = 'text'
    m_display_name.name = 'display_name'
    m_form.appendChild(display_name_l)
    m_form.appendChild(m_display_name)

    var isAdmin_l = document.createElement("label")
    isAdmin_l.innerText = "Admin"
    var m_isAdmin = document.createElement("input")
    m_isAdmin.type = 'checkbox'
    m_isAdmin.name = 'isAdmin'
    m_form.appendChild(isAdmin_l)
    m_form.appendChild(m_isAdmin)

    var color_picker_l = document.createElement("label")
    color_picker_l.innerText = "Badge color"
    var m_color_pick = document.createElement('input')
    m_color_pick.type = 'color'
    m_color_pick.name = 'color'
    m_color_pick.class = 'badge-color'
    m_form.appendChild(color_picker_l)
    m_form.appendChild(m_color_pick)

    var m_submit_btn = document.createElement('button')
    m_submit_btn.className = 'modal-close'
    m_submit_btn.innerText = "Update"

    m_form.appendChild(m_submit_btn)
}

function showUserEditModal(data) {
    var json_data = JSON.parse(data)

    // Elements of 'frags.js'
    var user_edit_frag = new DocumentFragment();
    user_edit_frag.appendChild(btnClose)
    user_edit_frag.appendChild(m_form)

    m_user_id.value = json_data['id']
    m_username.value = json_data['username']
    m_display_name.value = json_data['display_name']
    m_isAdmin.value = json_data['isAdmin']
    m_color_pick.value = json_data['color']

    modal_content.appendChild(user_edit_frag)

    modal.style.display = "flex";
}

export function StatusPopup() {
    let statusPopup = new DocumentFragment()

    let popupContainer = document.createElement('div') 
        popupContainer.id = "user-status-popup"
        popupContainer.className = "status-popup"
    
    let popupContent = document.createElement('div')
        popupContent.className = "status-popup-content"

    let onlineButton = document.createElement('button')
        onlineButton.id = "status-online"
        onlineButton.classList = "btn-fw btn-inv"
    let onlineBadge = document.createElement('img')
        onlineBadge.src = "static/images/online.png"
        onlineBadge.className = "status-badge"

    let awayButton = document.createElement('button')
        awayButton.id = "status-online"
        awayButton.classList = "btn-fw btn-inv"
    let awayBadge = document.createElement('img')
        awayBadge.src = "static/images/away.png"
        awayBadge.className = "status-badge"

    let busyButton = document.createElement('button')
        busyButton.id = "status-online"
        busyButton.classList = "btn-fw btn-inv"
    let busyBadge = document.createElement('img')
        busyBadge.src = "static/images/busy.png"
        busyBadge.className = "status-badge"

    onlineButton.addEventListener('click', () => {
        let data = {
            "event": "user_update",
            "status": "online"
        }
        ylk.Self.status = "online"
        ylk.Sock.postMessage(data)
        elementsHide(popupContainer)
    })

    awayButton.addEventListener('click', () => {
        let data = {
            "event": "user_update",
            "status": "away"
        }
        ylk.Self.status = "away"
        ylk.Sock.postMessage(data)
        elementsHide(popupContainer)
    })

    busyButton.addEventListener('click', () => {
        let data = {
            "event": "user_update",
            "status": "busy"
        }
        ylk.Self.status = "busy"
        ylk.Sock.postMessage(data)
        elementsHide(popupContainer)
    })

    onlineButton.appendChild(onlineBadge)
    onlineButton.append("Online")

    awayButton.appendChild(awayBadge)
    awayButton.append("Away")

    busyButton.appendChild(busyBadge)
    busyButton.append("Busy")

    popupContent.appendChild(onlineButton)
    popupContent.appendChild(awayButton)
    popupContent.appendChild(busyButton)

    popupContainer.appendChild(popupContent)
    statusPopup.appendChild(popupContainer)

    return statusPopup

};

export function elementsHide(target) {
    // var modal = document.getElementById("modal-box");
    // var search_popup = document.getElementById("search-popup");
    // if (modal != null) {
    //     modal.style.display = 'none'
    // }
    // if (search_popup != null) {
    //     search_popup.style.display = 'none'
    // }
    let popup = document.getElementById("user-status-popup")
    let search_popup = document.getElementById("search-popup")
    let modal = document.getElementById("modal-box")
    if (popup && target == popup) {
        popup.style.display = "none";
        popup.remove()
    }
    if (search_popup && target == search_popup) {
        document.querySelector('#searchInput').value = ''
        search_popup.style.display = "none";
        // search_popup.remove()
    }
    if (modal && target == modal) {
        modal.style.display = "none";
        modal.remove()
    }

}

window.onclick = (ev) => {
    elementsHide(ev.target)

}



// * Modal content elements

function NewSubmitButton(text) {
    let btnSubmit = document.createElement('button')
    btnSubmit.className = 'modal-close'
    btnSubmit.innerText = text
    return btnSubmit
}

// * Chat Buttons

export function NewHashtagText(name) {
    var icon_hashtag = document.createElement('i')
    icon_hashtag.classList = 'fa-solid fa-hashtag'
    var channel_text = document.createElement('span')
    channel_text.innerText = name
    return { icon_hashtag, channel_text }

}

export function NewBubbleText(name) {
    var icon_bubble = document.createElement('i')
    icon_bubble.classList = 'fa-solid fa-message'
    var bubble_text = document.createElement('span')
    bubble_text.innerText = name
    return { icon_bubble, bubble_text }
}

export function NewDMIcon(id) {
    var btnNewDm = document.createElement('button')
    btnNewDm.addEventListener('submit', function (event) {
        newDm(id)
    })
    var envelopeIcon = document.createElement('i')
    envelopeIcon.classList = 'fa-regular fa-envelope'
    return envelopeIcon
}

export function NewResultHeader(hdr_title) {
    let fragment = new DocumentFragment()

    let container = document.createElement('div')
    container.className = "side-item-container"
    container.id = "search-container"

    let section_header = document.createElement('div')
    section_header.className = "section-header"

    let title_text = document.createElement('span')
    title_text.className = "title"
    title_text.innerText = hdr_title

    section_header.appendChild(title_text)
    container.appendChild(section_header)
    fragment.appendChild(container)

    return fragment
}

export function WhiteSpacer() {
    let spacer = document.createElement('hr')
    spacer.className = 'white-border'

    return spacer
}

export function MessageRow(last_msg, user_id, type_message, display_name, text, isAdmin, color, _time, message_id) {
    let message = new DocumentFragment()

    var message_row = document.createElement('div')
    message_row.id = message_id

    if (type_message == "server-message") {
        var row_spacer = document.createElement('hr')
        row_spacer.classList = "white-border spacer"

        var text_message = document.createElement('span')
        text_message.className = "text"
        text_message.innerText = message

        var row_spacer = document.createElement('hr')
        row_spacer.className = "white-border"

        message_row.className = "server-message"
        message_row.appendChild(row_spacer)
        message_row.appendChild(text_message)
        message_row.appendChild(row_spacer)
        return message_row
    } else {
        var userDisplayName = document.createElement('span')
        userDisplayName.className = "username"
        userDisplayName.innerText = display_name

        var msg = document.createElement('span')
        msg.className = "message"
        msg.innerText += text

        let time, unix_time
        if (_time != "") {
            unix_time = Date.parse(_time)
            time = new Date(_time)
        } else {
            unix_time = Date.now();
            time = new Date()
        }

        var locale_date = time.toLocaleDateString()
        var locale_time = time.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })

        var message_time = document.createElement('span')
        message_time.innerText = locale_date + ' - ' + locale_time
        message_time.className = "timestamp"

        if (last_msg != null || last_msg != undefined) {
            var lm_from = last_msg.from
            var lm_unix_time = Date.parse(last_msg.time)

            if (lm_from == user_id && unix_time < (lm_unix_time + 600000) && type_message != "server_message") {
                var l_container = document.createElement('span')
                l_container.className = "timestamp-avatar"
                l_container.innerText = locale_time
                l_container.display = 'hidden'
                // l_container.className = "avatar"
                userDisplayName.style.display = 'none'
                message_time.style.display = 'none'
            } else {
                var l_container = document.createElement('img')
                l_container.className = "avatar"
                l_container.src = "/static/data/user_avatars/" + user_id + "/avatar.png"
                l_container.style.borderColor = color
            }
        } else {
            if (lm_from == user_id && unix_time > (lm_unix_time + 600000) && type_message != "server_message") {
                var l_container = document.createElement('span')
                l_container.innerText = locale_time
                userDisplayName.style.display = 'none'
                message_time.style.display = 'none'
            } else {
                var l_container = document.createElement('img')
                l_container.className = "avatar"
                l_container.src = "/static/data/user_avatars/" + user_id + "/avatar.png"
                l_container.style.borderColor = color
            }
        }
    }
    message_row.appendChild(l_container)
    message_row.appendChild(userDisplayName)
    message_row.appendChild(message_time)
    message_row.appendChild(msg)
    message_row.className = "message-row"
    message.append(message_row)
    return message
}


export function noResultRow() {
    var user_row = document.createElement('div')
    user_row.classList = 'profile-row'

    var text = document.createElement("label")
    text.innerText = "No results."
    user_row.appendChild(text)

    return user_row
}

export function UserButton(userData, containerNode) {
    let _btn = new DocumentFragment()
    var btn = document.createElement('button')
    btn.addEventListener('click', (ev) => {
        let data = Chat.Open("dm", "#", userData.user_id, ylk)
        if (data === null) {
            elementsHide(containerNode)
            return
        }
        Chat.Change(data)
    })
    var user_avatar = document.createElement('img')
    user_avatar.className = 'avatar'
    user_avatar.src = 'static/data/user_avatars/' + userData.user_id + '/avatar.png'

    user_avatar.style.borderColor = userData.color

    var status_dot = document.createElement('img')
    status_dot.className = 'status-badge'
    status_dot.src = 'static/images/' + userData.status + '.png'

    var username = document.createElement('span')
    username.className = 'username'
    username.innerText = userData.display_name

    if (containerNode != undefined && containerNode != null) { //&& (sel_ctr != undefined && sel_ctr != null)) {
        btn.id = 'user-profile-' + userData.user_id
        btn.classList = 'profile-row btn-fw btn-inv'

    } else {
        btn.id = userData.user_id
        btn.classList = 'profile-row btn-fw btn-inv'
        btn.addEventListener("click", function (input) {
            input.value = ""
            document.querySelector("#search-popup").style.display = 'none'
            let user = document.createElement('span')
            user.innerText = userData.user_id
        })
    }
    btn.appendChild(user_avatar)
    btn.appendChild(status_dot)
    btn.appendChild(username)
    var dmIcon = NewDMIcon(userData.user_id)
    btn.appendChild(dmIcon)
    _btn.appendChild(btn)

    return _btn
}