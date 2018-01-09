var HttpClient = function () {
    this.get = function (aUrl, aCallback) {
        var anHttpRequest = new XMLHttpRequest();
        anHttpRequest.onreadystatechange = function () {
            if (anHttpRequest.readyState == 4 && anHttpRequest.status == 200)
                aCallback(anHttpRequest.responseText);
        };

        anHttpRequest.open("GET", aUrl, true);
        anHttpRequest.send(null);
    }
};

function incrementShareCount(source, count) {
    var sharesCountDom = document.getElementById("shares-count");
    var newCount = count + parseInt(sharesCountDom.innerHTML);
    // I share each one, so only display if other people shared too
    if (newCount >= 2) {
        var sharesBoxDom = document.getElementById("shares-box");
        sharesBoxDom.style.display = "block";
        sharesCountDom.innerHTML = newCount;
    }
    if (console) {
        console.log(source + " shares: " + count);
    }
}

var client = new HttpClient();

client.get('https://graph.facebook.com/?id=' + window.location.href, function (response) {
    var fb = JSON.parse(response);
    var shareCount = fb.share.share_count;
    incrementShareCount("fb", shareCount)
});

client.get('/lkdcnt?url=' + window.location.href, function (response) {
    var lkd = JSON.parse(response);
    var shareCount = lkd.count;
    incrementShareCount("ldk", shareCount)
});

client.get('https://public.newsharecounts.com/count.json?url=' + window.location.href, function (response) {
    var twtr = JSON.parse(response);
    var shareCount = twtr.count;
    incrementShareCount("twtr", shareCount)
});



setInterval(function() {
    if (document.hasFocus()) {
        client.get("/ping?id={{.TrackingId}}"); // .TrackingId will be filled in via the go templating
    }
}, 5000);