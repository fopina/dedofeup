# -*- coding: utf-8 -*-
import mechanize
from HTMLParser import HTMLParser

BASE_URL = 'https://sigarra.up.pt/feup/pt/assd_tlp_geral.func_view?pct_val_cod=999999'
LOGIN_URL = 'https://sigarra.up.pt/feup/pt/web_page.inicial'


class FEUP():
    def __init__(self, username, password):
        self.cj = mechanize.LWPCookieJar()
        self._postloadinit()
        self.username = username
        self._password = password

    def _postloadinit(self):
        self.br = mechanize.Browser()
        self.br.set_handle_robots(False)
        useragent = 'Mozilla/5.0 (X11; U; Linux i686; en-US; rv:1.9.0.1) Gecko/2008071615 Fedora/3.0.1-1.fc9 Firefox/3.0.1'
        self.br.addheaders = [('User-agent', useragent)]
        self.br.set_cookiejar(self.cj)
        self.parser = MyHTMLParser()

    def login(self):
        try:
            res = self.br.open(BASE_URL)
            if res.read().find('class="border0"> Terminar sess\xe3o') > -1:
                return True
        except:
            pass

        self.br.open(LOGIN_URL)
        self.br.select_form(nr=0)
        if 'p_address' not in self.br:
            # wrong form - not login
            return False

        self.br['p_user'] = self.username
        self.br['p_pass'] = self._password

        res = self.br.submit()
        if res.read().find('class="border0"> Terminar sess\xe3o') > -1:
            return True
        else:
            return False

    def dedo(self, month=None, year=None):
        url = BASE_URL

        if month:
            url += '&pv_mes=' + str(month)

        if year:
            url += '&pv_ano=' + str(year)

        res = self.br.open(url)
        self.parser.feed(res.read().decode('Latin1'))

        # dirty cleanup - this should probably be done some other way

        days = self.parser.days

        for date in days:
            day = days[date]
            for marca in ['marca am', 'marca pm']:
                tmp = []
                for (i,val) in enumerate(day.get(marca,[])):
                    if i < 2:
                        tmp.append(val)
                        continue

                    if i % 2 == 0:
                        if val == '---':
                            break

                    tmp.append(val)
                day[marca] = tmp

        return days

    def __getstate__(self):
        odict = self.__dict__.copy() # copy the dict since we change it
        del odict['br']              # remove filehandle entry
        del odict['parser']
        return odict

    def __setstate__(self, dict):
        self.__dict__.update(dict)   # update attributes
        self._postloadinit()


class MyHTMLParser(HTMLParser):
    def __init__(self):
        HTMLParser.__init__(self)
        self.all_data = []
        self.day = {}
        self.days = {}
        self.daycol = 0
        self.dayinfo = None
        self.ignore_rest = False

    def handle_starttag(self, tag, attrs):
        if self.ignore_rest:
            return

        if tag == 'tr':
            class_name = self._getattr(attrs,'class')
            if class_name in ['dia-fds', 'dia-normal', 'dia-feriado', 'dia-atual']:
                self.day['class'] = [ class_name ]

        if self.day:
            if tag == 'td':
                class_name = self._getattr(attrs,'class')
                if class_name in [
                    'data k', 'saldo-d negativo', 'saldo-d', 'saldo-a',
                    'saldo-a negativo', 'marca am', 'marca pm',
                    'marca am aprovado', 'marca pm aprovado'
                    ]:
                    self.dayinfo = class_name

    def handle_endtag(self, tag):
        if tag == 'tr':
            if self.day:
                self.days[self.day['data k'][0]] = self.day
                self.day = {}
        elif tag == 'td':
            self.dayinfo = None
        elif tag == 'table':
            self.ignore_rest = True

    def handle_data(self, data):
        if self.dayinfo:
            tmp = self.dayinfo

            if tmp == 'marca am aprovado':
                tmp = 'marca am'
            elif tmp == 'marca pm aprovado':
                tmp = 'marca pm'
            elif tmp == 'saldo-d negativo':
                tmp = 'saldo-d'
            elif tmp == 'saldo-a negativo':
                tmp = 'saldo-a'

            self.day[tmp] = self.day.get(tmp,[]) + [ data ]

    def _getattr(self, attrs, name):
        for (key,val) in attrs:
            if key == name:
                return val
        return None


if __name__ == '__main__':
    import sys
    p = FEUP(sys.argv[1], sys.argv[2])
    if p.login():
        print p.dedo(month=3, year=2014)
    else:
        print 'Failed to login!'
