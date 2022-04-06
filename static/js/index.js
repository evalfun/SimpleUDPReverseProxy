var serverlist;
var clientlist;
var connlist;

var get_connection_stat = function (code) {
  switch (code) {
    case 0:
      return "初始化";
    case 1:
      return "连接客户端";
    case 3:
      return "连接服务器";
    case 4:
      return "就绪";
    case 5:
      return "连接成功";
    case 20:
      return "断开连接";
  }
}

var get_client_stat = function (code) {
  switch (code) {
    case 0:
      return "初始化";
    case 1:
      return "连接服务器";
    case 2:
      return "连接成功";
    case 20:
      return "关闭";
  }
}

function bytesToSize(bytes) {
  if (bytes === 0) return '0 B';
  var k = 1000, // or 1024
      sizes = ['B', 'KB', 'MB', 'GB', 'TB', 'PB', 'EB', 'ZB', 'YB'],
      i = Math.floor(Math.log(bytes) / Math.log(k));

 return (bytes / Math.pow(k, i)).toPrecision(3) + ' ' + sizes[i];
}


var list_server = function () {
  $.ajax({
    type: "GET",
    url: "/api/server/list",
    success: function (resp) {
      serverlist = resp;
      tpl = $.templates('#jsr-serverlist');
      final = tpl.render({ data: resp });
      var a = document.getElementById("div_serverlist_table");
      a.innerHTML = final;
    }
  })
}
var client_instance_id;
var list_client = function () {
  $.ajax({
    type: "GET",
    url: "/api/client/list",
    success: function (resp) {

      for (ins_id in resp) {
        resp[ins_id]['Stat'] = get_client_stat(resp[ins_id]['Stat']);
        if (resp[ins_id]["ConnectionStat"] != null) {
          for (i = 0; i < resp[ins_id]["ConnectionStat"].length; i++) {
            if (resp[ins_id]["ConnectionStat"][i]["ConnectStat"] == 5) {
              resp[ins_id]["RemoteName"] = resp[ins_id]["ConnectionStat"][i]["RemoteName"]
              resp[ins_id]["RemoteAddr"] = resp[ins_id]["ConnectionStat"][i]["RemoteAddr"]
              resp[ins_id]["RemoteOtherData"] = resp[ins_id]["ConnectionStat"][i]["RemoteOtherData"]
              break
            }
          }
        } else {
          resp[ins_id]["RemoteName"] = "";
          resp[ins_id]["RemoteAddr"] = "";
          resp[ins_id]["RemoteOtherData"] = "";

        }
        if (resp[ins_id]["ServerAddrList"] != null) {
          ServerAddrList = "";
          for (i = 0; i < resp[ins_id]["ServerAddrList"].length; i++) {
            ServerAddrList = ServerAddrList + resp[ins_id]["ServerAddrList"][i] + '\n';
          }
          resp[ins_id]["ServerAddrList"] = ServerAddrList;
        }
      }
      clientlist = resp;
      tpl = $.templates('#jsr-clientlist');
      final = tpl.render({ data: resp });
      var a = document.getElementById("div_clientlist_table");
      a.innerHTML = final;
    }
  })
}
var server_instance_id;
var list_server_conn = function () {
  $.ajax({
    type: "POST",
    url: "/api/server/connection",
    dataType: "json",
    data: JSON.stringify({ InstanceID: server_instance_id }),
    success: function (resp) {
      if (resp == null) {
        connlist = [];
      } else {
        for (i = 0; i < resp.length; i++) {
          resp[i]["ConnectStat"] = get_connection_stat(resp[i]["ConnectStat"]);
          resp[i]["DropRate"] = parseInt((1 - resp[i]["RecvPacketCount"] / resp[i]["RecvPacketSN"]) * 100);
          resp[i]["SendSize"] = bytesToSize(resp[i]["SendSize"]);
          resp[i]["RecvSize"] = bytesToSize(resp[i]["RecvSize"]);
        }
        connlist = resp;
      }

      tpl = $.templates('#jsr-connlist');
      final = tpl.render({ data: resp });
      var a = document.getElementById("div_server_conn_list");
      //console.log(final);
      a.innerHTML = final;
    }
  })
}

var list_client_conn = function () {
  $.ajax({
    type: "POST",
    url: "/api/client/connection",
    dataType: "json",
    data: JSON.stringify({ InstanceID: client_instance_id }),
    success: function (resp) {
      if (resp == null) {
        connlist = [];
      } else {
        for (i = 0; i < resp.length; i++) {
          resp[i]["ConnectStat"] = get_connection_stat(resp[i]["ConnectStat"]);
          resp[i]["DropRate"] = parseInt((1 - resp[i]["RecvPacketCount"] / resp[i]["RecvPacketSN"]) * 100);
          resp[i]["SendSize"] = bytesToSize(resp[i]["SendSize"]);
          resp[i]["RecvSize"] = bytesToSize(resp[i]["RecvSize"]);
        }
        connlist = resp;
      }
      tpl = $.templates('#jsr-connlist');
      final = tpl.render({ data: resp });
      var a = document.getElementById("div_client_conn_list");
      //console.log(final);
      a.innerHTML = final;
    }
  })
}

var list_server_session = function () {
  $.ajax({
    type: "POST",
    url: "/api/server/session",
    dataType: "json",
    data: JSON.stringify({ InstanceID: server_instance_id }),
    success: function (resp) {
      connlist = resp;
      tpl = $.templates('#jsr-session-list');
      session_list = [];
      for (addr in resp) {
        if (resp[addr] == null) {
          continue;
        }
        for (i = 0; i < resp[addr].length; i++) {
          session_list.push({
            RemoteAddr: addr,
            TargetAddr: resp[addr][i]['TargetAddr'],
            SendBytes: bytesToSize(resp[addr][i]['SendBytes']),
            RecvBytes: bytesToSize(resp[addr][i]['RecvBytes']),
            ClosedTime: resp[addr][i]['ClosedTime'],
          })
        }
      }
      console.log(session_list);
      final = tpl.render({ data: session_list });
      var a = document.getElementById("div_server_session");
      //console.log(final);
      a.innerHTML = final;
    }
  })
}

var list_client_session = function () {
  $.ajax({
    type: "POST",
    url: "/api/client/session",
    dataType: "json",
    data: JSON.stringify({ InstanceID: client_instance_id }),
    success: function (resp) {
      connlist = resp;
      tpl = $.templates('#jsr-session-list');
      session_list = [];
      for (addr in resp) {
        if (resp[addr] == null) {
          continue;
        }
        for (i = 0; i < resp[addr].length; i++) {
          session_list.push({
            RemoteAddr: addr,
            TargetAddr: resp[addr][i]['TargetAddr'],
            SendBytes: bytesToSize(resp[addr][i]['SendBytes']),
            RecvBytes: bytesToSize(resp[addr][i]['RecvBytes']),
            ClosedTime: resp[addr][i]['ClosedTime'],
          })
        }
      }
      console.log(session_list);
      final = tpl.render({ data: session_list });
      var a = document.getElementById("div_client_session");
      //console.log(final);
      a.innerHTML = final;
    }
  })
}

var show_conn_info = function (id) {
  var instance = M.Modal.getInstance(document.getElementById("modal_connection_info"));
  tpl = $.templates('#jsr-connection-info');
  final = tpl.render(connlist[id]);
  var a = document.getElementById("div_server_conn_info");
  a.innerHTML = final;
  instance.open();
  M.updateTextFields();
  M.textareaAutoResize($('#textarea2_'));

}

var show_server_info = function (server_id) {
  server_instance_id = server_id;
  tpl = $.templates('#jsr-serverinfo');
  final = tpl.render(serverlist[server_id]);
  var a = document.getElementById("div_server_info");
  a.innerHTML = final;

  var instance = M.Modal.getInstance(document.getElementById("modal_server_info"));
  instance.open();

  $('select').formSelect();
  M.updateTextFields();
  M.textareaAutoResize($('#input_server_otherdata'));
  $('.tabs').tabs();
  tab_ins = M.Tabs.getInstance(document.getElementById("tab_server_info"));
  tab_ins.select("server_info");
}

var show_client_info = function (client_id) {
  client_instance_id = client_id;
  tpl = $.templates('#jsr-clientinfo');
  final = tpl.render(clientlist[client_id]);
  var a = document.getElementById("div_client_info");
  a.innerHTML = final;

  var instance = M.Modal.getInstance(document.getElementById("modal_client_info"));
  instance.open();

  $('select').formSelect();
  M.updateTextFields();
  M.textareaAutoResize($('#input_client_otherdata'));
  M.textareaAutoResize($('#input_client_remoteaddrlist'));
  M.textareaAutoResize($('#input_client_remoteotherdata'));
  //$('.tabs').tabs();
  tab_ins = M.Tabs.getInstance(document.getElementById("tab_client_info"));
  tab_ins.select("client_info");
}


var show_create_server = function () {
  var instance = M.Modal.getInstance(document.getElementById("modal_create_server"));

  instance.open();
  $('select').formSelect();
  M.updateTextFields();
}
var show_create_client = function () {
  var instance = M.Modal.getInstance(document.getElementById("modal_create_client"));

  instance.open();
  $('select').formSelect();
  M.updateTextFields();
}

var create_server = function () {
  instance_id = parseInt($("#in_create_server_ins_id").val());
  local_port = parseInt($("#in_create_server_local_port").val());
  buf_size = parseInt($("#in_create_server_buf_size").val());
  password = $("#in_create_server_password").val();
  session_timeout = parseInt($("#in_create_server_session_timeout").val());
  save_closed_session = parseInt($("#in_create_server_save_close").val());
  target = $("#in_create_server_target").val();
  stun_server = $("#in_create_server_stun").val();
  target_version = document.getElementById("in_create_server_target_version").value;
  crypt_method = document.getElementById("in_create_server_crypt_method").value;
  encrypt_header_only = document.getElementById("in_create_server_eho").checked;
  hash_header_only = document.getElementById("in_create_server_hho").checked;
  local_name = $("#in_create_server_local_name").val();
  other_data = $("#in_create_server_other_data").val();
  var data = {
    InstanceID: instance_id,
    LocalPort: local_port,
    BufSize: buf_size,
    Target: target,
    TargetIPVersion: target_version,
    SessionTimeout: session_timeout,
    SaveClosedSession: save_closed_session,
    Password: password,
    CryptMethod: crypt_method,
    EncryptHeaderOnly: encrypt_header_only,
    HashHeaderOnly: hash_header_only,
    LocalName: local_name,
    OtherData: other_data,
    StunServer: stun_server,
  };
  //console.log(data);
  $.ajax({
    type: "POST",
    url: "/api/server/create",
    contentType: 'application/json',
    data: JSON.stringify(data),
    success: function (resp) {
      M.toast({ html: '创建成功' });
      list_server();
    },
    error: function (xhr, status, error) {
      show_info("创建失败", xhr.responseText);
    }
  })
}

var create_client = function () {
  instance_id = parseInt($("#in_create_client_ins_id").val());
  local_port = parseInt($("#in_create_client_local_port").val());
  listener_port = parseInt($("#in_create_client_listener_port").val());
  buf_size = parseInt($("#in_create_client_buf_size").val());
  password = $("#in_create_client_password").val();
  session_timeout = parseInt($("#in_create_client_session_timeout").val());
  save_closed_session = parseInt($("#in_create_client_save_close").val());
  target = $("#in_create_client_target").val();
  stun_server = $("#in_create_client_stun").val();
  target_version = document.getElementById("in_create_client_target_version").value;
  crypt_method = document.getElementById("in_create_client_crypt_method").value;
  encrypt_header_only = document.getElementById("in_create_client_eho").checked;
  hash_header_only = document.getElementById("in_create_client_hho").checked;
  local_name = $("#in_create_client_local_name").val();
  other_data = $("#in_create_client_other_data").val();
  server_addr = $("#in_create_client_remote_addr").val().split("\n");
  var data = {
    InstanceID: instance_id,
    LocalPort: local_port,
    ListenerPort: listener_port,
    BufSize: buf_size,
    Target: target,
    TargetIPVersion: target_version,
    SessionTimeout: session_timeout,
    SaveClosedSession: save_closed_session,
    Password: password,
    CryptMethod: crypt_method,
    EncryptHeaderOnly: encrypt_header_only,
    HashHeaderOnly: hash_header_only,
    LocalName: local_name,
    OtherData: other_data,
    StunServer: stun_server,
    CompressType: 0,
    ServerAddr: server_addr,
  };
  //console.log(data);
  $.ajax({
    type: "POST",
    url: "/api/client/create",
    contentType: 'application/json',
    data: JSON.stringify(data),
    success: function (resp) {
      M.toast({ html: '创建成功' });
      list_client();
    },
    error: function (xhr, status, error) {
      show_info("创建失败", xhr.responseText);
    }
  })
}

var connect_client = function () {
  $.ajax({
    type: "POST",
    url: "/api/server/connect",
    contentType: 'application/json',
    data: JSON.stringify({
      InstanceID: server_instance_id,
      ClientAddr: $("#in_client_conn").val(),
    }),
    success: function (resp) {
      M.toast({ html: '创建连接成功' });
      list_server_conn();
    },
    error: function (xhr, status, error) {
      show_info("失败", xhr.responseText);
    }
  });
}

var restart_client_connection = function () {
  $.ajax({
    type: "POST",
    url: "/api/client/connection/restart",
    contentType: 'application/json',
    data: JSON.stringify({
      InstanceID: client_instance_id,
    }),
    success: function (resp) {
      M.toast({ html: '成功重启客户端连接' });
      list_client();
    },
    error: function (xhr, status, error) {
      show_info("失败", xhr.responseText);
    }
  });
}

var update_client_serveraddr = function () {

  $.ajax({
    type: "POST",
    url: "/api/client/serveraddr/update",
    contentType: 'application/json',
    data: JSON.stringify({
      InstanceID: client_instance_id,
      ServerAddrList: $("#input_client_remoteaddrlist").val().split("\n")
    }),
    success: function (resp) {
      M.toast({ html: '成功更新服务端地址列表' });
    },
    error: function (xhr, status, error) {
      show_info("失败", xhr.responseText);
    }
  });
}

var delete_client = function () {
  $.ajax({
    type: "POST",
    url: "/api/client/delete",
    contentType: 'application/json',
    data: JSON.stringify({
      InstanceID: client_instance_id,
    }),
    success: function (resp) {
      M.toast({ html: '删除成功' });
      list_client();
    },
    error: function (xhr, status, error) {
      show_info("失败", xhr.responseText);
    }
  });
}

var delete_server = function () {
  $.ajax({
    type: "POST",
    url: "/api/server/delete",
    contentType: 'application/json',
    data: JSON.stringify({
      InstanceID: server_instance_id,
    }),
    success: function (resp) {
      M.toast({ html: '删除成功' });
      list_server();
    },
    error: function (xhr, status, error) {
      show_info("失败", xhr.responseText);
    }
  });
}

var save_config = function () {
  $.ajax({
    type: "GET",
    url: "/api/config/save",
    contentType: 'application/json',
    data: JSON.stringify({
      InstanceID: client_instance_id,
    }),
    success: function (resp) {
      M.toast({ html: '保存成功' });
      list_server();
    },
    error: function (xhr, status, error) {
      show_info("失败", xhr.responseText);
    }
  });
}


var show_info = function (title, content) {
  $("#modal_info_title").text(title);
  $("#modal_info_content").text(content);
  var instance = M.Modal.getInstance(document.getElementById("modal_info"));
  instance.open();
}

$(document).ready(function () {
  $('.modal').modal();
  $('.tabs').tabs();
  list_client();
});