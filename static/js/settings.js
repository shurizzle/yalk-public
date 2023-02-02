import { ready } from "./utils.js"

export function profile(userData) {
        let _profile = new DocumentFragment()

        // * Profile settings
        const profileBox = document.createElement('div')
        profileBox.classList = "box box-shadow box-flex-grow-2"
        
        const pictureBox = document.createElement('div')
        pictureBox.classList = "box box-shadow box-flex-grow-2"

        // const title = document.createElement('span')
        // title.className = "box-title"
        // title.innerText = "Users"

        const titlePicture = document.createElement('span')
        titlePicture.className = "box-title"
        titlePicture.innerText = "Profile Picture"

        const profileForm = document.createElement('form')
        profileForm.id = "user-profile"

        const labelDisplayName = document.createElement('label')
        labelDisplayName.innerText = "Display Name"

        const labelColorPicker = document.createElement('label')
        labelColorPicker.innerText = "Color"

        const inputColorPicker = document.createElement('input')
        inputColorPicker.type = "color"
        inputColorPicker.id = "color_pick"
        inputColorPicker.name = "color_pick"
        inputColorPicker.autocomplete = "off"
        inputColorPicker.value = userData.color

        const inputDisplayName = document.createElement('input')
        inputDisplayName.type = "text"
        inputDisplayName.id = "display_name"
        inputDisplayName.name = "display_name"
        inputDisplayName.autocomplete = "off"
        inputDisplayName.value = userData.display_name

        const btnUpdate = document.createElement('input')
        btnUpdate.type = "submit"
        btnUpdate.className = "wide-button"
        btnUpdate.value = "Update"
    
        // * Profile Picture settings
        
        const profilePicture = document.createElement('img')
        profilePicture.id = "profile-avatar"
        profilePicture.className = "avatar-profile"
        profilePicture.src = "static/data/user_avatars/" + userData.user_id + "/avatar.png"
        
        const uploadPicture = document.createElement('input')
        uploadPicture.type = "file"
        uploadPicture.id = "picture"
        uploadPicture.name = "picture"

        const btnSubmit = document.createElement('input')
        btnSubmit.id = "picture-submit"
        btnSubmit.type = "submit"
        btnSubmit.class = "wide-button"
        btnSubmit.value = "Update"

        btnSubmit.addEventListener('click', () => {
            const file = document.querySelector('#picture').files[0]
            const reader = new FileReader();
            reader.addEventListener('load', () => {
                const queueData = {
                    'event': 'user_update'
                }
                ylk.Queued.push(queueData)
                const data = {
                    'event': 'data_image',
                    'data': btoa(reader.result)
                }
                ylk.Sock.postMessage(data)
            }, false)
            if (file) {
                reader.readAsBinaryString(file)
            }
        })

        profileForm.appendChild(labelDisplayName)
        profileForm.appendChild(inputDisplayName)
        profileForm.appendChild(labelColorPicker)
        profileForm.appendChild(inputColorPicker)
        profileForm.appendChild(btnUpdate)

        profileBox.appendChild(profileForm)

        // container.appendChild(box)
        pictureBox.appendChild(titlePicture)
        pictureBox.appendChild(profilePicture)
        pictureBox.appendChild(uploadPicture)
        pictureBox.appendChild(btnSubmit)


        // _profile.appendChild(title)
        _profile.appendChild(profileBox)
        _profile.appendChild(pictureBox)

        
        return _profile

    }


// ready(() => {

// })

// function addRow() {
//     if (row_added == true) {
//         return
//     }
//     const table = document.getElementById('users-table')

//     var row = table.insertRow(-1)
//     var id_cell = row.insertCell(-1)
//     var username_cell = row.insertCell(-1)
//     var password_cell = row.insertCell(-1) 
//     var admin_cell = row.insertCell(-1)
//     var action_cell = row.insertCell(-1)
    
//     // action_cell.class

//     // id_cell.innerText = '🆕'
//     // username_cell.innerHTML = '<input class="max_width_250" type="text" form="new-user" id="username-new" name="username-new" autocomplete="off"></input>'
//     // password_cell.innerHTML = '<input ="max_width_250" type="password" form="new-user" id="password-new" name="password-new" autocomplete="new-password"></input>'
//     // admin_cell.innerHTML = '<select formclass="new-user" id="admin" name="admin" autocomplete="off"><option value="true">Yes</option><option value="false">No</option></select>'
//     // action_cell.innerHTML = '<button type="submit" class="btn-action" id="button-new" form="new-user"><i class="fa-solid fa-check"></i></button>'

//     row_added = true
//     var new_user_form = document.createElement("form")
//     new_user_form.id = 'new-user'
//     new_user_form.action = '/admin/user/add'
//     new_user_form.method = 'POST'
//     // new_user_form.addEventListener('submit', function(e) {
//     //     button_new = $('#button-new').get(0)
//     //     button_new.innerHTML = '<i class="fas fa-spinner fa-pulse"></i>'
//     //     button_new.disabled = true
//     //     e.preventDefault(); // avoid to execute the actual submit of the form.
//     //     $.ajax({
//     //         type: $(this).attr('method'),
//     //         url: $(this).attr('action'),
//     //         xhrFields: {
//     //             withCredentials: true
//     //         },
//     //         data: $(this).serialize(), // serializes the form's elements.
//     //         success: function(response) {
//     //             username_new = $('#username-new').get(0)
//     //             isAdmin_new = $('#admin').get(0)
//     //             id_cell.innerText = response
//     //             username_cell.innerHTML = ''
//     //             username_cell.innerText = username_new.value
//     //             password_cell.innerHTML = ''
//     //             password_cell.innerText = '**********'
//     //             admin_cell.innerHTML = ''
//     //             admin_cell.innerText = isAdmin_new.value
//     //             action_cell.innerHTML = '<button class="btn-action" onclick="UserDelete(' + response + ')"><i class="fa-solid fa-trash"></i></button>'
//     //             row_added = false
//     //         },
//     //         error: function(xhr){
//     //             console.log('Error adding user: ' + xhr)
//     //         }
//     //     });
//     // });
//     action_cell.appendChild(new_user_form)
// }

