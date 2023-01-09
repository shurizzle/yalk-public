function addRow() {
    if (row_added == true) {
        return
    }
    const table = document.getElementById('users-table')

    var row = table.insertRow(-1)
    var id_cell = row.insertCell(-1)
    var username_cell = row.insertCell(-1)
    var password_cell = row.insertCell(-1) 
    var admin_cell = row.insertCell(-1)
    var action_cell = row.insertCell(-1)
    
    action_cell.class

    id_cell.innerText = '🆕'
    username_cell.innerHTML = '<input class="max_width_250" type="text" form="new-user" id="username-new" name="username-new" autocomplete="off"></input>'
    password_cell.innerHTML = '<input class="max_width_250" type="password" form="new-user" id="password-new" name="password-new" autocomplete="new-password"></input>'
    admin_cell.innerHTML = '<select form="new-user" id="admin" name="admin" autocomplete="off"><option value="true">Yes</option><option value="false">No</option></select>'
    action_cell.innerHTML = '<button type="submit" class="btn-action" id="button-new" form="new-user"><i class="fa-solid fa-check"></i></button>'

    row_added = true
    var new_user_form = document.createElement("form")
    new_user_form.id = 'new-user'
    new_user_form.action = '/admin/user/add'
    new_user_form.method = 'POST'
    new_user_form.addEventListener('submit', function(e) {
        button_new = $('#button-new').get(0)
        button_new.innerHTML = '<i class="fas fa-spinner fa-pulse"></i>'
        button_new.disabled = true
        e.preventDefault(); // avoid to execute the actual submit of the form.
        $.ajax({
            type: $(this).attr('method'),
            url: $(this).attr('action'),
            xhrFields: {
                withCredentials: true
            },
            data: $(this).serialize(), // serializes the form's elements.
            success: function(response) {
                username_new = $('#username-new').get(0)
                is_admin_new = $('#admin').get(0)
                id_cell.innerText = response
                username_cell.innerHTML = ''
                username_cell.innerText = username_new.value
                password_cell.innerHTML = ''
                password_cell.innerText = '**********'
                admin_cell.innerHTML = ''
                admin_cell.innerText = is_admin_new.value
                action_cell.innerHTML = '<button class="btn-action" onclick="UserDelete(' + response + ')"><i class="fa-solid fa-trash"></i></button>'
                row_added = false
            },
            error: function(xhr){
                console.log('Error adding user: ' + xhr)
            }
        });
    });
    action_cell.appendChild(new_user_form)
}