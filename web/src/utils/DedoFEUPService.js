import axios from 'axios';

const SERVER_URL = {
    '_': 'https://func.skmobi.com/function/dedofeup',
    'dev': 'https://func.skmobi.com/function/dedofeup-dev',
    'local': 'http://localhost:9999',
}

var days = [];

var token = null;

export function setToken(newToken) {
    token = newToken
    if (typeof window !== "undefined") {
        if (newToken == null) {
            localStorage.removeItem("token")
        }
        else {
            localStorage.setItem("token", newToken)
        }
    }
}

export function getToken() {
	if (typeof window !== "undefined") {
		token = localStorage.getItem("token")
	}
	return token
}
export function isLoggedIn() {
    return getToken()
}

export function refreshWithLogin(username, password) {
    return axios.post(getServerEndpoint(), {'username': username, 'password': password})
                .then(response => response.data)
}

export function refreshWithToken(token) {
    return axios.post(getServerEndpoint(), {'token': token})
                .then(response => response.data)
}

export function setDays(newDays) {
    days = newDays.reverse()
    // remove all "future" days
    console.log(days)
    while (days.length && (days[0].Type !== "atual")) {
        days.shift()
    }

    if (typeof window !== "undefined") {
    	localStorage.setItem("days", JSON.stringify(days))
    }
}

export function getDays() {
	if (typeof window !== "undefined") {
		days = JSON.parse(localStorage.getItem("days"))
	}
    return days
}

function getServerEndpoint() {
    var c = '_'
    if (typeof window !== "undefined") {
        c = localStorage.getItem("apiServer")
	}
    return SERVER_URL[c] || SERVER_URL['_']
}

export function setServerEndpoint(server) {
    if (typeof window !== "undefined") {
    	localStorage.setItem("apiServer", server)
    }
}
