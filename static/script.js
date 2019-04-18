$('#save-crm').on("submit", function(e) {
    e.preventDefault();
    let formData = formDataToObj($(this).serializeArray());
    $(this).find('button.btn').addClass('disabled');
    $(this).find(".material-icons").addClass('animate');
    $("form :input").prop("disabled", true);
    send(
        $(this).attr('action'),
        formData,
        function (data) {
            sessionStorage.setItem("createdMsg", data.message);

            document.location.replace(
                location.protocol.concat("//").concat(window.location.host) + data.url
            );
        }
    )
});

$("#but-settings").on("click", function(e) {
    e.preventDefault();
    $(this).addClass('disabled');
    $(this).find(".material-icons").addClass('animate');
    send(
        $(this).attr('data-action'),
        {
            client_id: $(this).attr('data-clientID'),
            lang: $("select#lang").find(":selected").text(),
            currency: $("select#currency").find(":selected").val(),
            timezone: $("select#timezone").find(":selected").val()
        },
        function (data) {
            M.toast({
                html: data.msg,
                displayLength: 1000,
                completeCallback: function(){
                    $(document).find('#but-settings').removeClass('disabled');
                    $(document).find(".material-icons").removeClass('animate');
                }
            });
        }
    )
});

$("#save").on("submit", function(e) {
    e.preventDefault();
    let formData = formDataToObj($(this).serializeArray());
    $(this).find('button.btn').addClass('disabled');
    $(this).find(".material-icons").addClass('animate');
    $("form :input").prop("disabled", true);
    send(
        $(this).attr('action'),
        formData,
        function (data) {
            M.toast({
                html: data.msg,
                displayLength: 1000,
                completeCallback: function(){
                    $(document).find('button.btn').removeClass('disabled');
                    $(document).find(".material-icons").removeClass('animate');
                    $("form :input").prop("disabled", false);
                }
            });
        }
    )
});

function send(url, data, callback) {
    $.ajax({
        url: url,
        data: JSON.stringify(data),
        type: "POST",
        success: callback,
        error: function (res) {
            if (res.status >= 400) {
                M.toast({
                    html: res.responseJSON.error,
                    displayLength: 1000,
                    completeCallback: function(){
                        $(document).find('button.btn').removeClass('disabled');
                        $(document).find(".material-icons").removeClass('animate');
                        $("form :input").prop("disabled", false);
                    }
                })
            }
        }
    });
}

function formDataToObj(formArray) {
    let obj = {};
    for (let i = 0; i < formArray.length; i++){
        obj[formArray[i]['name']] = formArray[i]['value'];
    }
    return obj;
}

$( document ).ready(function() {
    $('select').formSelect();
    M.Tabs.init(document.getElementById("tab"));

    let createdMsg = sessionStorage.getItem("createdMsg");
    if (createdMsg) {
        setTimeout(function() {
            M.toast({
                html: createdMsg
            });
            sessionStorage.removeItem("createdMsg");
        }, 1000);
    }
});
