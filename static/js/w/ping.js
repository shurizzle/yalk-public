function ping(){
    var xhttp = new XMLHttpRequest()
    xhttp.withCredentials = true
    xhttp.onreadystatechange = function() {
        if (this.readyState == 4 && this.status == 200) {
            postMessage("ok")
            this.responseText;
        }
    }
    xhttp.open("POST", "/client-ping", true);
    xhttp.setRequestHeader("Content-type", "multipart/x-www-form-urlened");
    xhttp.send();
}

setInterval(() => {
    ping()
 }, 30000)