let conn, _data, main, root, aw, ah;
window.onload = function () {
    // todo: реализовать корректно закрытие вебсокета
    main = document.getElementById('main');
    root = document.documentElement;

    prepareEnviroment();

    if (window["WebSocket"]) {
        let protocol = 'ws';
        if (document.location.host.substr(0, 2) !== '0.') {
            // probably Production
            protocol += 's'
        }
        conn = new WebSocket(protocol + '://' + document.location.host + "/ws");
        conn.onclose = function (evt) {
            console.warn('Connection closed.');
        };
        conn.onmessage = function (evt) {
            // console.debug('Sent: ' + evt.data.substr(0, 100));
            if (evt.data === 'ERR') {
                console.error('Там у них ERR');
                return;
            }
            let data = JSON.parse(evt.data);
            _data = data;
            let res = {
                'DrawMapCmd': ifDrawMap,
                'DrawPlant': ifDrawPlant,
                'DrawHerbivoreAnimal': isDrawHerbivoreAnimal,
                'DrawPredatoryAnimal': isDrawPredatoryAnimal,
                'InfoAbout': isInfoAbout,
                'MoveMe': isMoveMe,
                'MustDie': isMustDie,
                'Bue': isBue,
            }[data['OnCmd']](data);

            if (!res) {
                console.error('Detected problem ' + Object.entries(data));
            }
        };
        conn.onopen = function (evt) {
            console.info('WebSocket has opened just');
            conn.send(JSON.stringify({'Cmd': 'entity'}));
            // conn.send(JSON.stringify({'Cmd': 'init', 'Id': -1}));
        };

        conn.onerror = function (error) {
            console.error(error.message);
        };
    } else {
        console.info('Your browser does not support WebSockets.');
    }
};

function prepareEnviroment() {
    root.style.setProperty('--set-border-entity', '1px');
    document.getElementById('turnBorderItem').checked = true;
    document.getElementById('turnBorderItem').onchange = changeSetBorderEntity;
}

function changeSetBorderEntity() {
    root.style.setProperty('--set-border-entity', {
        '0px': '1px',
        '1px': '0px',
    }[root.style.getPropertyValue('--set-border-entity')]);
}

function isBue(data) {
    alert(data['Reason']);
    return true;
}

function isMustDie(data) {
    let ent = document.getElementById('_go_' + data['Id']);
    if (ent === null) return false;
    $('#_go_' + data['Id']).fadeOut(1000, function () {
        ent.remove()
    });
    return true;
}

function _(op, a, b) {
    if (op === 'A')
        return parseInt(a.replace('px', '')) + b + 'px';
    if (op === 'B') {
        let pref = parseInt(a.replace('px', ''));
        if (pref < 0) return '0px';
        if (pref >= b) return b + 'px';
        return a;
    }
}

function isMoveMe(data) {
    function chooseColor(c) {
        if (c < 0.2) return 'chartreuse';
        if (c < 0.4) return 'khaki';
        if (c < 0.6) return 'coral';
        if (c < 0.8) return 'crimson';
        return 'maroon';
    }
    // todo: поворачивать когда куда нить идёт

    let ent = document.getElementById('_go_' + data['IdObj']);
    if (ent === null) return false;
    ent.style.left = _('A', ent.style.left, data['ChangeX']);
    ent.style.top = _('A', ent.style.top, data['ChangeY']);
    let h = ent.getElementsByClassName('health-progress')[0];
    h.style.width = (100 - parseInt(data['Hunger'] * 100)) + '%';
    h.style['background-color'] = chooseColor(data['Hunger']);
    return true;
}

function isInfoAbout(data) {
    if (data['Class'] === 'Plant') {
        alert('Это растение');
        return true;
    } else if (data['Class'] === 'HerbivoreAnimal') {
        alert('Это травоядное животное.');
        if (data['Target'])
            alert('Он охотятится');
        alert('Он голоден на ' + data['Hunger'] + 'из 100');
        return true;
    } else if (data['Class'] === 'PredatoryAnimal') {
        alert('Это хищное животное.');
        if (data['Target'])
            alert('Он охотятится');
        alert('Он голоден на ' + data['Hunger'] + 'из 100');
        return true;
    }
}

function addEntity(data) {
    let ent = document.createElement('div');
    ent.id = '_go_' + data['Id'];
    ent.className = 'entity';
    main.insertBefore(ent, null);
    ent.style.top = data['Top'] + 'px';
    ent.style.left = data['Left'] + 'px';
    ent.__godata__ = {'id': data['Id']};
    ent.onclick = function () {
        conn.send(JSON.stringify({
            'Cmd': 'info',
            'Id': data['Id'],
        }))
    };
    return ent
}

function addHealthCheck(p) {
    let health = document.createElement('div');
    health.className = 'health-check';
    p.insertBefore(health, null);
    let progress = document.createElement('div');
    progress.className = 'health-progress';
    health.insertBefore(progress, null);
}

function isDrawPredatoryAnimal(data) {
    let halimal = addEntity(data);
    addHealthCheck(halimal);
    halimal.style.background = 'url(/static/imgs/panimal.png)';
    halimal.style['background-size'] = '100%';
    return true;
}

function isDrawHerbivoreAnimal(data) {
    let halimal = addEntity(data);
    addHealthCheck(halimal);
    halimal.style.background = 'url(/static/imgs/hanimal.png)';
    halimal.style['background-size'] = '100%';
    return true;
}

function ifDrawPlant(data) {
    let plant = addEntity(data);
    plant.style.background = {
        0: 'url(/static/imgs/plant_type1.png)',
        1: 'url(/static/imgs/plant_type2.png)',
        2: 'url(/static/imgs/plant_type3.png)',
        3: 'url(/static/imgs/plant_type4.png)',
        4: 'url(/static/imgs/plant_type5.png)',
        5: 'url(/static/imgs/plant_type6.png)',
    }[data['Type']];
    plant.style['background-size'] = '100%';
    return true;
}

function ifDrawMap(data) {
    let indent_top = 0, indent_left = 0;
    let width_item = 10, height_item = 10;
    aw = data['Gap'].length * width_item - 30;
    ah = data['Gap'][0].length * height_item - 30;

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
