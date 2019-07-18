#!/usr/bin/env python
# -*- coding: utf-8 -*-
from flask import Flask, session, redirect, url_for, request, render_template, flash
from feup import FEUP
from functools import wraps
import pickle
from datetime import datetime, timedelta

# Flask FLASH categories
FLASH_ERROR = 'danger'
FLASH_SUCCESS = 'success'

app = Flask(__name__)


def requires_login():
    def wrapper(f):
        @wraps(f)
        def wrapped(*args, **kwargs):
            if 'profile' not in session:
                return redirect(url_for('login'))
            return f(*args, **kwargs)
        return wrapped
    return wrapper


@app.route('/', methods=['GET'])
@requires_login()
def index():
    profile = pickle.loads(session['profile'])
    try:
        back = int(request.args.get('back', 0))
        if back < 0:
            back = 0
    except ValueError:
        back = 0

    try:
        force_reload = int(request.args.get('reload', 0))
    except ValueError:
        force_reload = 0

    date = datetime.now() - timedelta(days=back)

    date_str = '%d-%02d-%02d' % (date.year, date.month, date.day)

    data = session.get('data', {})

    if (date_str not in data) or force_reload:
        try:
            session['data'] = profile.dedo(month=date.month, year=date.year)
        except:  # TODO failed to login? try to re-login?
            try:
                if profile.login():
                    session['data'] = profile.dedo(month=date.month, year=date.year)
                    session['profile'] = pickle.dumps(profile)
                else:
                    return redirect(url_for('logout'))
            except:  # something wrong happened, let's redirect to logout to force re-auth
                return redirect(url_for('logout'))
        data = session.get('data', {})

    return render_template('query.html', date=date_str, info=data.get(date_str, {}), back=back)


@app.route('/login', methods=['GET', 'POST'])
def login():
    try:
        if request.method == 'POST':
            profile = FEUP(request.form['username'], request.form['password'])
            if profile.login():
                session['profile'] = pickle.dumps(profile)
                session.permanent = True
            else:
                flash('Login Failed', FLASH_ERROR)

            return redirect(url_for('index'))

        return render_template('login.html')
    except:
        raise


@app.route('/logout')
def logout():
    session.clear()
    return redirect(url_for('index'))

# set the secret key.  keep this really secret:
app.secret_key = 'A0Zr98j/3yX R~XHH!jmN]LWX/,?R1'

if __name__ == '__main__':
    app.run(debug=True)
