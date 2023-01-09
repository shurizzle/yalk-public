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

    let title = document.createElement("label")
    let chatType;

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
            btnSubmit = newSubmitButton("Create")
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
            var btnSubmit = newSubmitButton("Create")

            name.type = 'text'
            name.name = 'dms_name'
            name.addEventListener('keyup', function (e) {
                searchUserAdd(e, sel_ctr)
            })

            form.appendChild(title)
            form.appendChild(lbl_sel)
            form.appendChild(sel_ctr)
            form.appendChild(lbl_name)
            form.appendChild(name)
            form.appendChild(btnSubmit)
            break

        // case "user":
        //     var form = newFormTemplate("/dm/create", "POST", "New Direct Message", newDm)
        //     var lbl_name = document.createElement("label")
        //     lbl_name.innerText = "Username"
        //     form.appendChild(lbl_name)
        //     var name = document.createElement("input")
        //     name.type = 'text'
        //     name.name = 'dms_name'
        //     form.appendChild(name)
        //     btnSubmit = newSubmitButton("Create")
        //     form.appendChild(btnSubmit)
        //     return form
    }
    modal_content.append(form)
    modal.append(modal_content)

    return modal
    // modal.addEventListener("click", function (e) {
    //     modal.style.display = "none";
    //     modal_content.innerHTML = ''
    // })
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

    var is_admin_l = document.createElement("label")
    is_admin_l.innerText = "Admin"
    var m_is_admin = document.createElement("input")
    m_is_admin.type = 'checkbox'
    m_is_admin.name = 'is_admin'
    m_form.appendChild(is_admin_l)
    m_form.appendChild(m_is_admin)

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
    m_is_admin.value = json_data['is_admin']
    m_color_pick.value = json_data['color']

    modal_content.appendChild(user_edit_frag)

    modal.style.display = "flex";
}

export function showStatusPopup() {
    document.getElementById("user-status-popup").style.display = "block";
};

export function elementsHide() {
    var modal = document.getElementById("modal-box");
    var search_popup = document.getElementById("search-popup");
    if (modal != null) {
        modal.style.display = 'none'
    }
    if (search_popup != null) {
        search_popup.style.display = 'none'
    }
}

window.onclick = function (event) {
    let popup = document.getElementById("user-status-popup")
    let search_popup = document.getElementById("search-popup")

    let modal = document.getElementById("modal-box")

    let remove = false;
    switch (event.target) {
        case popup:
            popup.style.display = "none";
            remove = true
            break

        case search_popup:
            var input = document.getElementById("searchInput")
            input.value = ""
            search_popup.style.display = "none";
            remove = true
            break

        case modal:
            modal.style.display = "none";
            remove = true
            break
    }
    if (remove === true){event.target.remove()}
    // if (event.target == popup || event.target == search_popup) {
    //     popup.style.display = "none";
    // }
}

// function newForm(action, method, form_title, cb, chatType, data) {
//     let form = document.createElement("form")
//     form.id = "modal-form"
//     form.action = action
//     form.method = method
//     form.className = 'flex-column'

//     form.addEventListener('submit', function (e) {
//         e.preventDefault(); // avoid to execute the actual submit of the form.
//         cb(chatType, e.target.value, [], ylk)
//         let modal = document.getElementById("modal-box")
//         if (modal != null) { modal.remove() } //TODO: Move in app.js on the return

//     });
//     var title = document.createElement("label")
//     title.innerText = form_title
//     form.appendChild(title)
//     return form
// }

// * Modal content elements

function newSubmitButton(text) {
    var btnSubmit = document.createElement('button')
    btnSubmit.className = 'modal-close'
    btnSubmit.innerText = text
    return btnSubmit
}

// * Chat Buttons

export function newHashtagText(name) {
    var icon_hashtag = document.createElement('i')
    icon_hashtag.classList = 'fa-solid fa-hashtag'
    var channel_text = document.createElement('span')
    channel_text.innerText = name
    return { icon_hashtag, channel_text }

}

export function newBubbleText(name) {
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

export function elementsMessageRow(last_msg, user_id, type_message, display_name, message, is_admin, color, _timestamp, message_id) {
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
        msg.innerText += message

        if (_timestamp != "") {
            var unix_time = Date.parse(_timestamp)
            var timestamp = new Date(_timestamp)
        } else {
            var unix_time = Date.now();
            var timestamp = new Date()
        }

        var locale_date = timestamp.toLocaleDateString()
        var locale_time = timestamp.toLocaleTimeString([], { hour: '2-digit', minute: '2-digit' })

        var message_timestamp = document.createElement('span')
        message_timestamp.innerText = locale_date + ' - ' + locale_time
        message_timestamp.className = "timestamp"

        if (last_msg != null) {
            var lm_from = last_msg.from
            var lm_unix_time = Date.parse(last_msg.timestamp)

            if (lm_from == user_id && unix_time < (lm_unix_time + 600000) && type_message != "server_message") {
                var l_container = document.createElement('span')
                l_container.className = "timestamp-avatar"
                l_container.innerText = locale_time
                l_container.display = 'hidden'
                // l_container.className = "avatar"
                userDisplayName.style.display = 'none'
                message_timestamp.style.display = 'none'
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
                message_timestamp.style.display = 'none'
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
    message_row.appendChild(message_timestamp)
    message_row.appendChild(msg)
    message_row.className = "message-row"
    return message_row
}

// export function NewDmButton(id, users, Users, CurrentUser) {
//     let dm_button = document.createElement('button')
//     dm_button.id = id
//     dm_button.className = 'side-item sidebar-item'
//     const self_index = users.indexOf(CurrentUser)
//     if (self_index > -1) {
//         users.splice(self_index, 1)
//     }
//     var users_display_names = []
//     users.forEach(element => {
//         users_display_names.push(ylk.Users[element].display_name)
//     })
//     let bubble_items = newBubbleText(users_display_names.toString())

//     dm_button.appendChild(bubble_items.icon_bubble)
//     dm_button.appendChild(bubble_items.bubble_text)

//     dm_button.addEventListener('click', function (e) {
//         changeDm(id)
//     })
//     return dm_button
// }

export function noResultRow() {
    var user_row = document.createElement('div')
    user_row.classList = 'profile-row'

    var text = document.createElement("label")
    text.innerText = "No results."
    user_row.appendChild(text)

    return user_row
}