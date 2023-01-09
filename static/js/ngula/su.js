// All admin commands, this file will be explicitely loaded 
// if the user is admin

function updateUser(id) {
    // Open modal.js modal
    $.ajax({
        type: 'get',
        url: '/user?' + $.param({'id': id}),
        xhrFields: {
            withCredentials: true
        },
        success: function(response) {
            json_data = JSON.parse(response)
            showModal("user", "placeholder_cb", ylk, json_data)
        },
        error: function(xhr){
            console.log(xhr)
        }
    })
}

function UserDelete(id){
    $.ajax({
        type: 'delete',
        url: '/admin/user/delete?' + $.param({'id': id}),
        xhrFields: {
            withCredentials: true
        },
        // data: $(this).serialize(), // serializes the form's elements. // ? Should it be disabled??
        success: function(response) {
            $('#user_row_'+id).remove()
        },
        error: function(xhr){
            console.log(xhr)
        }
    });
};