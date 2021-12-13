AT&T

<_start>:
  0:  48 b8 2f 62 69 6e 2f    movabs $0x68732f6e69622f,%rax         # '/bin/sh'
  7:  73 68 00
  a:  50                      push   %rax
  b:  48 89 e7                mov    %rsp,%rdi                      # command line
  e:  48 31 f6                xor    %rsi,%rsi
 11:  48 31 d2                xor    %rdx,%rdx
 14:  48 31 c0                xor    %rax,%rax
 17:  b0 3b                   mov    $0x3b,%al
 19:  0f 05                   syscall                               # execve
 1b:  48 31 c0                xor    %rax,%rax
 1e:  b0 3c                   mov    $0x3c,%al
 20:  0f 05                   syscall                               # exit
