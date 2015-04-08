$(document).ready(function () {
	var fill_ol = function(data, title) {
		title = typeof title !== 'undefined' ? title : '';
		if (title) {
			new_title = $("<p></p>").text(title);
		} else {
			new_title = null
		}
		var new_ol = $("<ol>");
		for (var i=data.length-1; i>=0; i--) {
			new_ol.append($("<li>").append($('<span class="song-title">').text(data[i].Title)).append($('<span class="song-url">').text(data[i].Hash)));
		}
		if (new_title) {
			return $(new_title).add(new_ol);
		}
		return new_title + new_ol;
	}
	var messagebox = $('<span id="message"/>');
	messagebox.message = function (text, delay) {
		if (typeof delay !== 'undefined') {
			messagebox.text(text);
			setTimeout(function () { messagebox.text(""); }, delay);
		} else {
			messagebox.text(text);
		}
	}
	var music_text = $('<input type="text" id="music_text" value=""/>')
		.focusout(function () {
			setTimeout(function () { $("#music_text").focus(); }, 100);
		})
		.keyup(function (e) {
			if (e.keyCode == 13) $("#music_add").click();
		});
	var music_add = $('<input type="button" id="music_add" value="add music"/>').click(function () {
			messagebox.message("Adding to queue...");
			$.get("/addq?hash=" + music_text.val())
				.always(function (data) {
					music_text.val("");
					if (data.lastIndexOf("OK", 0) === 0) {
						getLists();
						messagebox.message("");
					} else {
						getLists();
						messagebox.message("unable to extract audio for this url", 2000);
					}
				});
		});
	var music_skip = $('<input type="button" id="music_skip" value="skip"/>').click(function () { $.get("/abrt").always(function () { getLists(); }); });
	var music_stop = $('<input type="button" id="music_stop" value="stop"/>').click(function () { $.get("/stop").always(function () { getLists(); }); });
	var music_resume = $('<input type="button" id="music_resume" value="resume"/>').click(function () { $.get("/resu").always(function () { getLists(); }); });
	var music_forever = $('<input type="button" value="forever"/>').click(function () { $.get("/loop").always(function () { getLists(); }); });
	var now_playing = $('<span id="nowtext"/>').text("Now playing: ");
	$("#now").append(messagebox);
	$("#now").append(music_skip);
	$("#now").append(music_stop);
	$("#now").prepend(music_forever);
	$("#now").append(music_resume);
	$("#now").append(now_playing);
	$("#now").append("<br/>");
	$("#now").append(music_text);
	$("#now").append(music_add);
	var nowtext = function(s) {
		$("#nowtext").text(s)
		document.title = s;
	};
	music_text.focus();
	var getLists = function () {
		$.get("/nowp", function (data) {
			if (data.lastIndexOf("STOPPED", 0) === 0) {
				nowtext("Player stopped");
				$("#music_resume").show();
				$("#music_forever").show();
				$("#music_skip").hide();
				$("#music_stop").hide();
			} else if (data.lastIndexOf("OK ", 0) === 0) {
				var title = data.substring(3);
				if (title !== "") {
					nowtext("Now playing: " + title);
					$("#music_resume").hide();
					$("#music_forever").show();
					$("#music_skip").show();
					$("#music_stop").show();
				} else {
					nowtext("Queue empty");
					$("#music_resume").hide();
					$("#music_forever").show();
					$("#music_skip").hide();
					$("#music_stop").show();
				}
			} else {
				// error
				$("#nowtext").text("Unknown/error");
				document.title = 'Unknown/error';
			}
		});
		$.getJSON("/list", function (data) {
			$("#q").html(fill_ol(data, "Play queue:"));
			$("#q ol li").prepend($('<input type="button" value="remove"/>').click(function (x) {
				$.get("/delq?hash=" + $(x.target).siblings(".song-url").text()).always(function (data) { getLists(); messagebox.message("removed stream: " + data, 1000); });
			}));
		});
		$.getJSON("/hist", function (data) {
			$("#h").html(fill_ol(data, "Play history:"));
			$("#h ol li").prepend($('<input type="button" value="remove"/>').click(function (x) {
				$.get("/delh?hash=" + $(x.target).siblings(".song-url").text()).always(function (data) { getLists(); messagebox.message("removed stream: " + data, 1000); });
			}));
			$("#h ol li").prepend($('<input type="button" value="requeue"/>').click(function (x) {
				$.get("/addq?hash=" + $(x.target).siblings(".song-url").text()).always(function (data) { getLists(); messagebox.message("requeued stream: " + data, 1000); });
			}));
		});
	}
	getLists();
	setInterval(getLists, 3000);
});
