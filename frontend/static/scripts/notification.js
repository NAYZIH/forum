const ws = new WebSocket("ws://" + window.location.host + "/ws");
 ws.onmessage = function(event) {
     const data = JSON.parse(event.data);
     if (data.type === "notification") {
         const countElem = document.getElementById("notification-count");
         if (countElem) {
             countElem.textContent = data.count > 0 ? data.count : "";
         }
     }
 };