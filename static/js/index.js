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

var get_crypt_method = function (code) {
  switch (code) {
    case 9:
      return "AES_GCM";
    case 7:
      return "NONE_SHA1";
    case 8:
      return "NONE_CRC32";
    default:
      return "未知";
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
      //console.log(session_list);
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


var show_create_client = function (ip, crypt) {

  tpl = $.templates('#jsr_in_create_client_target_version');
  final = tpl.render({ TargetIPVersion: ip });
  var a = document.getElementById("in_create_client_target_version");
  a.innerHTML = final;


  tpl = $.templates('#jsr_in_create_client_crypt_method');
  final = tpl.render({ EncryptMethod: crypt });
  var a = document.getElementById("in_create_client_crypt_method");
  a.innerHTML = final;


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
  //encrypt_header_only = document.getElementById("in_create_server_eho").checked;
  //hash_header_only = document.getElementById("in_create_server_hho").checked;
  local_name = $("#in_create_server_local_name").val();
  other_data = $("#in_create_server_other_data").val();
  tracker_config = {
    ServerID: $("#in_create_server_tracker_server_id").val(),
    UserID: $("#in_create_server_tracker_user_id").val(),
    ServerURL: $("#in_create_server_tracker").val(),
  }
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
    //EncryptHeaderOnly: encrypt_header_only,
    //HashHeaderOnly: hash_header_only,
    LocalName: local_name,
    OtherData: other_data,
    StunServer: stun_server,
    Tracker: tracker_config,
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
  //encrypt_header_only = document.getElementById("in_create_client_eho").checked;
  //hash_header_only = document.getElementById("in_create_client_hho").checked;
  local_name = $("#in_create_client_local_name").val();
  other_data = $("#in_create_client_other_data").val();
  server_addr = $("#in_create_client_remote_addr").val().split("\n");
  tracker_config = {
    ServerID: $("#in_create_client_tracker_server_id").val(),
    UserID: $("#in_create_client_tracker_user_id").val(),
    ServerURL: $("#in_create_client_tracker").val(),
  }
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
    //EncryptHeaderOnly: encrypt_header_only,
    //HashHeaderOnly: hash_header_only,
    LocalName: local_name,
    OtherData: other_data,
    StunServer: stun_server,
    CompressType: 0,
    ServerAddr: server_addr,
    Tracker: tracker_config,
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
    success: function (resp) {
      M.toast({ html: '保存成功' });
      list_server();
    },
    error: function (xhr, status, error) {
      show_info("失败", xhr.responseText);
    }
  });
}

var get_server_tracker = function () {
  $.ajax({
    type: "POST",
    url: "/api/server/tracker",
    contentType: 'application/json',
    data: JSON.stringify({
      InstanceID: server_instance_id,
    }),
    success: function (resp) {
      $("#in_server_tracker").val(resp.ServerURL);
      $("#in_server_tracker_server_id").val(resp.ServerID);
      $("#in_server_tracker_user_id").val(resp.UserID);
      $("#sp_server_tracker_stat").text(resp.Message);

      var target;
      if (serverlist[server_instance_id].Target == "") {
        target = "[Edit Me]";
      } else {
        target = serverlist[server_instance_id].Target
      }
      var client_tracker_url;
      if (resp.ServerURL.startsWith("wss")) {
        client_tracker_url = resp.ServerURL.replace("wss", "https")
      } else {
        client_tracker_url = resp.ServerURL.replace("ws", "http")
      }
      var tracker_config = {
        ServerID: resp.ServerID,
        UserID: "",
        ServerURL: client_tracker_url,
      };
      encrypt_method = get_crypt_method(serverlist[server_instance_id].EncryptMethod);
      var client_config = {
        InstanceID: 0,
        ListenerPort: 0,
        LocalPort: 0,
        BufSize: serverlist[server_instance_id].BufSize,
        Target: target,
        TargetIPVersion: serverlist[server_instance_id].TargetIPVersion,
        SessionTimeout: serverlist[server_instance_id].SessionTimeout,
        SaveClosedSession: serverlist[server_instance_id].SaveClosedSession,
        Password: serverlist[server_instance_id].Password,
        CryptMethod: encrypt_method,
        //EncryptHeaderOnly: serverlist[server_instance_id].EncryptHeaderOnly,
       // HashHeaderOnly: serverlist[server_instance_id].HashHeaderOnly,
        StunServer: serverlist[server_instance_id].StunServer,
        Tracker: tracker_config,
      }
      $("#in_server_fc_link").val("surp://" + Base64.encode(JSON.stringify(client_config)));
      M.updateTextFields();
      M.textareaAutoResize($('#in_server_fc_link'));
    },
    error: function (xhr, status, error) {
      show_info("失败", xhr.responseText);
    }
  });
}

var set_server_tracker = function () {
  data = {
    InstanceID: server_instance_id,
    ServerURL: $("#in_server_tracker").val(),
    ServerID: $("#in_server_tracker_server_id").val(),
    UserID: $("#in_server_tracker_user_id").val(),
  }
  $.ajax({
    type: "POST",
    url: "/api/server/tracker/set",
    contentType: 'application/json',
    data: JSON.stringify(data),
    success: function (resp) {
      M.toast({ html: '更新成功' });
      get_server_tracker();
    },
    error: function (xhr, status, error) {
      show_info("失败", xhr.responseText);
    }
  });
}


var get_client_tracker = function () {
  $.ajax({
    type: "POST",
    url: "/api/client/tracker",
    contentType: 'application/json',
    data: JSON.stringify({
      InstanceID: client_instance_id,
    }),
    success: function (resp) {
      $("#in_client_tracker").val(resp.ServerURL);
      $("#in_client_tracker_server_id").val(resp.ServerID);
      $("#in_client_tracker_user_id").val(resp.UserID);
      $("#sp_client_tracker_stat").text(resp.Message);


      target = clientlist[client_instance_id].Target;
      client_tracker_url = resp.ServerURL
      var tracker_config = {
        ServerID: resp.ServerID,
        UserID: "",
        ServerURL: client_tracker_url,
      };
      encrypt_method = get_crypt_method(clientlist[client_instance_id].CryptMethod);
      //console.log(clientlist[client_instance_id]);
      var client_config = {
        InstanceID: 0,
        ListenerPort: 0,
        LocalPort: 0,
        BufSize: clientlist[client_instance_id].BufSize,
        Target: target,
        TargetIPVersion: clientlist[client_instance_id].TargetIPVersion,
        SessionTimeout: clientlist[client_instance_id].SessionTimeout,
        SaveClosedSession: clientlist[client_instance_id].SaveClosedSession,
        Password: clientlist[client_instance_id].Password,
        CryptMethod: encrypt_method,
        //EncryptHeaderOnly: clientlist[client_instance_id].EncryptHeaderOnly,
        //HashHeaderOnly: clientlist[client_instance_id].HashHeaderOnly,
        StunServer: clientlist[client_instance_id].StunServer,
        Tracker: tracker_config,
      }
      $("#in_client_fc_link").val("surp://" + Base64.encode(JSON.stringify(client_config)));
      M.updateTextFields();
      M.textareaAutoResize($('#in_client_fc_link'));
    },
    error: function (xhr, status, error) {
      show_info("失败", xhr.responseText);
    }
  });
}

var set_client_tracker = function () {
  data = {
    InstanceID: client_instance_id,
    ServerURL: $("#in_client_tracker").val(),
    ServerID: $("#in_client_tracker_server_id").val(),
    UserID: $("#in_client_tracker_user_id").val(),
  }
  $.ajax({
    type: "POST",
    url: "/api/client/tracker/set",
    contentType: 'application/json',
    data: JSON.stringify(data),
    success: function (resp) {
      M.toast({ html: '更新成功' });
      //get_client_tracker();
    },
    error: function (xhr, status, error) {
      show_info("失败", xhr.responseText);
    }
  });
}

var create_client_by_link = function () {
  link = $("#in_create_client_fc_link").val();
  if (!link.startsWith("surp://")) {
    show_info("错误", "链接格式错误");
    return;
  }
  try {
    link = Base64.decode(link.substr(7))
  }
  catch (err) {
    show_info("解析链接错误", "base64: " + err.message);
    return
  }
  var client_config;
  try {
    client_config = JSON.parse(link);
  }
  catch (err) {
    show_info("解析链接错误", "json: " + err.message);
    return
  }
  $("#in_create_client_ins_id").val(client_config.InstanceID);
  $("#in_create_client_listener_port").val(client_config.ListenerPort);
  $("#in_create_client_local_port").val(client_config.LocalPort);
  $("#in_create_client_password").val(client_config.Password);
  $("#in_create_client_target").val(client_config.Target);
  //console.log(client_config.Target);
  $("#in_create_client_stun").val(client_config.StunServer);
  $("#in_create_client_local_name").val("_" + Math.random().toString(10).slice(-8));
  $("#in_create_client_tracker").val(client_config.Tracker.ServerURL);
  $("#in_create_client_tracker_server_id").val(client_config.Tracker.ServerID);

  $("#in_create_client_buf_size").val(client_config.BufSize);
  $("#in_create_client_session_timeout").val(client_config.SessionTimeout);

  //document.getElementById("in_create_client_hho").checked = client_config.HashHeaderOnly;
  //document.getElementById("in_create_client_eho").checked = client_config.EncryptHeaderOnly;

  $("#in_create_client_save_close").val(client_config.SaveClosedSession);
  show_create_client(client_config.TargetIPVersion, client_config.CryptMethod);
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