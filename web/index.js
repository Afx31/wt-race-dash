const socket = new WebSocket('ws://' + document.location.host + '/ws');

socket.onopen = function(event) {
  console.log('Connected to WebSocket server.');
};

socket.onmessage = function(event) {
  const data = JSON.parse(event.data);

  var rpmBar = document.getElementById('rpmbar');
  var rpmNum = document.getElementById('rpmNum');
  var speed = document.getElementById('speed');
  var gear = document.getElementById('gear');
  var voltage = document.getElementById('voltage');
  var iat = document.getElementById('iat');
  var ect = document.getElementById('ect');
  var tpsBar = document.getElementById('tpsbar');
  var tps = document.getElementById('tps');
  var map = document.getElementById('map');
  var lambdaRatio = document.getElementById('lambdaRatio');
  // var inj = document.getElementById('inj');
  // var ign = document.getElementById('ign');
  var oilTemp = document.getElementById('oilTemp');
  var oilPressure = document.getElementById('oilPressure');
  
  // Assign data to UI controls
  rpmBar.style.width = ((data.Rpm / 9000) * 100) + '%';

  // if (tpsBar.style.height !== data.tps + '%')
    tpsBar.style.height = data.Tps + '%';

  // if (rpmNum.textContent !== data.rpm)
    rpmNum.textContent = data.Rpm;
  
  // if (speed.textContent !== data.speed)
    speed.textContent = data.Speed;
  
  gear.textContent = data.Gear;
  voltage.textContent = data.Voltage;  
  iat.textContent = data.Iat;
  ect.textContent = data.Ect;
  tps.textContent = data.Tps;
  map.textContent = data.Map;
  lambdaRatio.textContent = data.LambdaRatio;
  // inj.textContent = data.inj;
  // ign.textContent = data.ign;
  oilTemp.textContent = data.OilTemp;
  oilPressure.textContent = data.OilPressure;
};

socket.onerror = function(error) {
  console.log('WebSocket error:', error);
};

socket.onclose = function(event) {
  console.log('WebSocket connection closed:', event);
};