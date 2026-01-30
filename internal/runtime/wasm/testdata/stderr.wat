;; A minimal WASI program that prints "error\n" to stderr
(module
  ;; Import WASI fd_write function
  (import "wasi_snapshot_preview1" "fd_write"
    (func $fd_write (param i32 i32 i32 i32) (result i32)))

  ;; Import WASI proc_exit function
  (import "wasi_snapshot_preview1" "proc_exit"
    (func $proc_exit (param i32)))

  ;; Memory (1 page = 64KB)
  (memory (export "memory") 1)

  ;; Data section: "error\n" at offset 8
  (data (i32.const 8) "error\n")

  ;; iov structure at offset 0:
  ;; [0-3]: pointer to string (8)
  ;; [4-7]: length of string (6)
  (data (i32.const 0) "\08\00\00\00\06\00\00\00")

  ;; _start function (WASI entry point)
  (func (export "_start")
    ;; Call fd_write(stderr=2, iovs=0, iovs_len=1, nwritten=16)
    (call $fd_write
      (i32.const 2)   ;; fd: stderr
      (i32.const 0)   ;; iovs pointer
      (i32.const 1)   ;; iovs length
      (i32.const 16)  ;; nwritten pointer
    )
    drop

    ;; Exit with code 0
    (call $proc_exit (i32.const 0))
  )
)
