function upload(form) {
  $form = $('#upload-form');
  fd = new FormData($form[0]);
  $.ajax(
    '/api/process',
    {
      type: 'post',
      processData: false,
      contentType: false,
      data: fd,
      dataType: "json",
      success: function(data) {
        if (data.result != true) {
          alert(data.message);
        } else {
          $("#result").attr("src", "data:image/png;base64," + data.body);
        }
      },
      error: function(XMLHttpRequest, textStatus, errorThrown) { }
    });
    return false;
}
