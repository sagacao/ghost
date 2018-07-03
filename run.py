#!/usr/bin/env python
# coding=utf-8

#nohup ./logapp &>app.log
import os
import sys
import time

user = 'saga'
process = 'ghost'


def get_pid():
    cmd = "ps -u %s | grep %s| grep -v grep |awk '{printf $1\" \"}'" % (user, process)
    pid = os.popen(cmd).read()
    if pid:
        return pid
    else:
        return 0
    pass

def start():
    cmd = "nohup ./%s > out.file 2>&1 &" % process
    os.popen(cmd)
    time.sleep(0.5)
    pid = get_pid()
    if pid != 0:
        print ">>  %s start success! " % process
    else:
        print ">>  %s start failed ... " % process
    pass

def stop():
    pid = get_pid()
    if pid:
        cmd = "kill -9 %s" % pid
        os.popen(cmd)
        time.sleep(0.2)
        pid = get_pid()
        if pid == 0:
            print ">>  %s stop success! " % process
        else:
            print ">>  %s stop failed ... " % process
    pass

def state():
    pid = get_pid()
    if pid == 0:
        print ">> %s not exit ..." % process
    else:
        print ">> %s pid: %s " % (process, pid)
    pass

def do_action(action):
    if action == 'start':
        start()
    elif action == 'stop':
        stop()
    elif action == 'restart':
        stop()
        start()
    else:
        state()
    pass


if __name__ == "__main__":
    opt = sys.argv
    print ">>>> -------------------------------------------"
    if len(opt) == 1:
        print "Usage: "
        print "python run.py [cmd:start/stop/restart]"
        exit()

    action = opt[1]
    do_action(action)
    print "<<<< -------------------------------------------"
