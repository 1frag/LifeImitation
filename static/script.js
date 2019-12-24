let conn, _data, main, aw, ah;
window.onload = function () {
    main = document.getElementById('main');
    if (window["WebSocket"]) {
        conn = new WebSocket("ws://" + document.location.host + "/ws");
        conn.onclose = function (evt) {
            console.warn('Connection closed.');
        };
        conn.onmessage = function (evt) {
            console.debug('Sent: ' + evt.data.substr(0, 100));
            if (evt.data === 'ERR') {
                console.error('Там у них ERR');
                return;
            }
            let data = JSON.parse(evt.data);
            _data = data;
            let res = {
                'DrawMap': ifDrawMap,
                'DrawPlant': ifDrawPlant,
                'DrawHerbivoreAnimal': isDrawHerbivoreAnimal,
                'InfoAbout': isInfoAbout,
            }[data['OnCmd']](data);

            if (!res) {
                console.error('Detected problem ' + Object.entries(data));
            }
        };
        conn.onopen = function (evt) {
            console.info('WebSocket has opened just');
            conn.send(JSON.stringify({'Cmd': 'init', 'Id': -1}));
        }
    } else {
        console.info('Your browser does not support WebSockets.');
    }
};

function isInfoAbout(data) {
    if (data['Class'] === 'Plant') {
        alert('Это растение');
        return true;
    } else if (data['Class'] === 'ResponsePlants') {
        alert('Это травоядное животное.');
        if (data['Target'])
            alert('Он охотятится');
        alert('Он голоден на ' + data['Hunger'] + 'из 100');
        return true;
    }
}

function isDrawHerbivoreAnimal(data) {
    let halimal = document.createElement('div');
    halimal.className = 'entity';
    main.insertBefore(halimal, null);
    halimal.style.top = (data['Top'] * aw) + 'px';
    halimal.style.left = (data['Left'] * ah) + 'px';
    halimal.style.background = 'url(/static/imgs/hanimal.png)';
    halimal.style['background-size'] = '100%';
    halimal.__godata__ = {'id': data['Id']};
    halimal.onclick = function () {
        conn.send(JSON.stringify({
            'Cmd': 'info',
            'Id': data['Id'],
        }))
    };
    return true;
}

function ifDrawPlant(data) {
    let plant = document.createElement('div');
    plant.className = 'entity';
    main.insertBefore(plant, null);
    plant.style.top = (data['Top'] * aw) + 'px';
    plant.style.left = (data['Left'] * ah) + 'px';
    plant.style.background = {
        0: 'url(/static/imgs/plant_type1.png)',
        1: 'url(/static/imgs/plant_type2.png)',
        2: 'url(/static/imgs/plant_type3.png)',
        3: 'url(/static/imgs/plant_type4.png)',
        4: 'url(/static/imgs/plant_type5.png)',
        5: 'url(/static/imgs/plant_type6.png)',
    }[data['Type']];
    plant.onclick = function () {
        conn.send(JSON.stringify({
            'Cmd': 'info',
            'Id': data['Id'],
        }))
    };
    plant.style['background-size'] = '100%';
    plant.__godata__ = {'id': data['Id']};
    return true;
}

function ifDrawMap(data) {
    let indent_top = 0, indent_left = 0;
    let width_item = 10, height_item = 10;
    aw = data['Gap'].length * width_item - 30;
    ah = data['Gap'][0].length * height_item - 30;
    conn.send(JSON.stringify({'Cmd': 'entity', 'Id': -1}));
    for (let i = 0; i < data['Gap'].length; i++) {
        for (let j = 0; j < data['Gap'][i].length; j++) {
            let e = document.createElement('div');
            e.className = 'item';
            e.style.background = {
                0: '#f0db7d', /*песок*/
                1: '#a2653e', /*земля*/
                2: '#3f9b0b', /*трава*/
                3: '#00bae4', /*река*/
            }[data['Gap'][i][j]];
            main.insertBefore(e, null);
            e.style.top = (indent_top + height_item * i) + 'px';
            e.style.left = (indent_left + width_item * j) + 'px';
        }
    }
    return true;
}
