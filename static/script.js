let conn, _data, main, root, aw, ah;
window.onload = function () {
    // todo: реализовать корректно закрытие вебсокета
    main = document.getElementById('main');
    root = document.documentElement;

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
            let res = ({
                'DrawPeople': isDrawPeople,
                'DrawMapCmd': isDrawMap,
                'DrawPlant': isDrawPlant,
                'DrawAnimal': isDrawAnimal,
                'InfoAbout': isInfoAbout,
                'MoveMe': isMoveMe,
                'MakeFence': isMakeFence,
                'MustDie': isMustDie,
                'DrawHouse': isDrawHouse,
                'Bue': isBue,
            }[data['OnCmd']] || err_detect)(data);

            if (!res) {
                failed[failed.length] = data;
                console.log(res);
                console.info(data);
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

let failed = [];

function isMakeFence(data) {
    let p1 = {
        'Left': Math.min(data['Point1']['Left'], data['Point2']['Left']),
        'Top': Math.min(data['Point1']['Top'], data['Point2']['Top']),
    };
    let p2 = {
        'Left': Math.max(data['Point1']['Left'], data['Point2']['Left']),
        'Top': Math.max(data['Point1']['Top'], data['Point2']['Top']),
    };
    let ent = document.createElement('div');
    ent.className = 'fence';
    main.insertBefore(ent, null);
    ent.style.top = p1['Top'] + 'px';
    ent.style.left = p1['Left'] + 'px';
    ent.style.width = p2['Left'] - p1['Left'] + 'px';
    ent.style.width = p2['Left'] - p1['Left'] + 'px';
    return ent
}

function err_detect(data) {
    failed[failed.length] = data;
    return false;
}

function isDrawHouse(data) {
    let house = addEntity(data);
    house.className += ' house';
    house.style.background = _url_('house');
    house.style['background-size'] = '100%';
}

function _url_(of) {
    return 'url(/static/imgs/' + of + '.png)'
}

function isDrawPeople(data) {
    let people = addEntity(data);
    addHealthCheck(people);
    let curl = '';
    if (5 < data['Age'] && data['Age'] < 18 && data['Gender'] === 'Female') {
        curl = _url_('girl')
    } else if (5 < data['Age'] && data['Age'] < 18) {
        curl = _url_('boy')
    } else if (data['Age'] >= 18 && data['Gender'] === 'Female') {
        curl = _url_('woman')
    } else if (data['Age'] >= 18) {
        curl = _url_('man')
    } else {
        curl = _url_('child')
    }
    people.style.background = curl;
    people.style['background-size'] = '100%';
    return true;
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
    failed[failed.length] = data;
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
    halimal.style.background = _url_('panimal');
    halimal.style['background-size'] = '100%';
    return true;
}

function isDrawAnimal(data) {
    let animal = addEntity(data);
    addHealthCheck(animal);
    animal.style.background = _url_({
        "Кролик": 'hanimal',
        "Волк": 'panimal',
        "Медведь": 'bear',
        "Зебра": 'zebra',
        "Лиса": 'fox',
        "Слон": 'elephant',
    }[data['Class']]);
    animal.style['background-size'] = '100%';
    animal.style['background-repeat'] = 'no-repeat';
    return true;
}

function isDrawPlant(data) {
    let plant = addEntity(data['Data']);
    plant.style.background = _url_({
        "Морковь": 'carrot',
        "Капуста": 'cabbage',
        "Кустарник": 'bush',
    }[data['Data']['Kind']]);
    plant.style['background-size'] = '100%';
    plant.style['background-repeat'] = 'no-repeat';
    return true;
}

function isDrawMap(data) {
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
