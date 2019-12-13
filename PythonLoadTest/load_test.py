import os
import sys
import requests
from faker import Faker
import random
import json
from datetime import datetime
import time
from threading import Thread

random.seed(datetime.now())
fake = Faker()
os.environ['NO_PROXY'] = '127.0.0.1'

HOST = "http://46.101.161.68:8080/"

LOGIN_TEMPLATE = """
 {{   
    "email": "{}",
    "password": "{}"
}}
"""

NEW_ACCOUNT_TEMPLATE = """
{{
    "nickname": "{0}",
    "email": "{1}",
    "password": "{2}",
    "sex": "{3}",
    "description": "{4}"
}}
"""



SEX = ["female", "male"]


def del_spaces(s):
    return "".join(s.split())

def random_hash_string(n):
    n%=16
    return str(random.getrandbits(128))[:n]

def create_new_accounts(n, file_name):
    accounts_file = open(file_name,"a")
    accounts = []
    good = 0
    for i in range(n):
        nick = random_hash_string(4) + fake.name() 
        if len(nick) > 20:
            nick = nick[:20]
        email = del_spaces(fake.name()+random_hash_string(4)) + "@gmail.com"
        password = fake.name()
        sex = SEX[bool(random.getrandbits(1))]
        description = fake.text()[:random.randint(1,200)]
        description = description.replace("\n", " ")
        payload = NEW_ACCOUNT_TEMPLATE.format(nick, email, password, sex, description)

        start = time. time()
        r = requests.post(HOST + "user/create", data=payload)
        end = time. time()
        print("Time elapssed: ", end-start)

        response = json.loads(r.text)
        
        if r.status_code != 200:
            print(response)
            print(payload)
        else:

            payload = LOGIN_TEMPLATE.format(email, password)
            r = requests.post(HOST + "user/login", data=payload)
            response = json.loads(r.text)
            token = response['access_token']
            data = {'email': email, 'password': password, 'token': "Bearer " + token}
            accounts.append(data)
            good += 1

    accounts_file.write(json.dumps(accounts, indent=4))
    print("Successful registrations: ", good)

def login_accounts_and_save_tokens():
    accounts = []
    with open(ACCOUNTS_FILE, "r") as fp:
        accounts = json.load(fp)
    
    for account in accounts:
        payload = LOGIN_TEMPLATE.format(account["email"], account["password"])
        r = requests.post(HOST + "user/login", data=payload)

        response = json.loads(r.text)
        if r.status_code != 200:
            print(response["message"])
        else:
            print(f"Logged in user: {account['email']}")
            account["token"] = "Bearer " + response["access_token"]
    accounts_file = open(ACCOUNTS_FILE,"w")
    accounts_file.write(json.dumps(accounts))

def measure_time(f):
    start = time. time()
    f()
    end = time.time()
    print("Time elapssed: ", end-start)

def get_lobbies(n, token):
    for i in range(n):
        start = time.time()

        headers = {'Authorization': "Bearer " + token}
        r = requests.get(HOST + "lobbies/results", headers=headers)

        end = time.time()
        print("Time elapssed: ", end-start)

    # data = json.loads(r.text)
    # print(json.dumps(data, indent=4))

def get_current_lobby(token):
    def wrapper():
        headers = {'Authorization': token}
        r = requests.get(HOST + "lobbies/my/current_lobby", headers=headers)
    measure_time(wrapper)

def get_my_profile(token):
    def wrapper():
        headers = {'Authorization': token}
        r = requests.get(HOST + "user/me", headers=headers)
    measure_time(wrapper)

def get_players(token):
    def wrapper():
        headers = {'Authorization': token}
        r = requests.get(HOST + "players", headers=headers)
    measure_time(wrapper)

    

def apply_tokens(func):
    def wrapper():
        accounts = []
        with open(ACCOUNTS_FILE, "r") as fp:
            accounts = json.load(fp)
        for acc in accounts:
            func(acc["token"])
    return wrapper





print("Threads: ", sys.argv[1])


threads = []
accounts_per_thread = int(sys.argv[2])
filename = sys.argv[3]
for i in range(int(sys.argv[1])):
    threads.append(Thread(target = create_new_accounts, args =(accounts_per_thread,f"{filename}_{i}.json",)))

for thread in threads:
    thread.start()

for thread in threads:
    thread.join()