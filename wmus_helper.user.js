// ==UserScript==
// @name        wmus helper
// @namespace   wmus
// @description Adds a "queue to wmus" button to youtube pages.
// @include     https://www.youtube.com/watch*
// @include     http://www.youtube.com/watch*
// @version     1
// @require     http://ajax.googleapis.com/ajax/libs/jquery/1.7.2/jquery.min.js
// @grant       GM_info
// ==/UserScript==

// NOTE:
// if one is logged into youtube, all urls are redirected
// to https, which means cross-site request that queues a
// video into wmus only works when wmus url is also https

function GM_main ($) {
  var wmus_uri = localStorage["wmus_uri"] || false;
  if (!wmus_uri) {
    $("#eow-title")
      .append('<input type="text" id="wmus_uri" value="" />')
      .append($('<input type="button" value="set wmus_uri" />')
        .click(function (x) {
	  localStorage["wmus_uri"] = $("#wmus_uri").val().replace(/\/+$/g, '');
	  location.reload();
	}));
  } else {
    $("#eow-title")
      .append($('<input type="button" value="add to queue" />')
        .click(function (x) {
	  $.get(wmus_uri + "/addq?hash=" + location.href)
	  .always(function (data) {
	    if (data.lastIndexOf("OK", 0) === 0) {
	      $(x.target).val("queued!");
	    } else {
	      $(x.target).val("ERROR: " + data);
	    }
	    setTimeout(function () { $(x.target).val("add to queue"); }, 1500);
	  });
	}))
      .append($('<input type="button" value="settings" />')
	.click(function (x) {
	  localStorage.removeItem("wmus_uri");
	  location.reload();
	}));
  }
    
}

if (typeof jQuery === "function") {
    GM_main (jQuery);
}
else {
    add_jQuery (GM_main, "2.1.2");
}

function add_jQuery (callbackFn, jqVersion) {
    var jqVersion   = jqVersion || "2.1.2";
    var D           = document;
    var targ        = D.getElementsByTagName ('head')[0] || D.body || D.documentElement;
    var scriptNode  = D.createElement ('script');
    scriptNode.src  = 'http://ajax.googleapis.com/ajax/libs/jquery/'
                    + jqVersion
                    + '/jquery.min.js'
                    ;
    scriptNode.addEventListener ("load", function () {
        var scriptNode          = D.createElement ("script");
        scriptNode.textContent  =
            'var gm_jQuery  = jQuery.noConflict (true);\n'
            + '(' + callbackFn.toString () + ')(gm_jQuery);'
        ;
        targ.appendChild (scriptNode);
    }, false);
    targ.appendChild (scriptNode);
}

if (!this.GM_getValue || (this.GM_getValue.toString && this.GM_getValue.toString().indexOf("not supported")>-1)) {
  this.GM_getValue=function (key,def) {
      return localStorage[key] || def;
  };
  this.GM_setValue=function (key,value) {
      return localStorage[key]=value;
  };
  this.GM_deleteValue=function (key) {
      return delete localStorage[key];
  };
}
