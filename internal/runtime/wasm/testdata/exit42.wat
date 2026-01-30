;; A minimal WASI program that exits with code 42
(module
  ;; Import WASI proc_exit function
  (import "wasi_snapshot_preview1" "proc_exit"
    (func $proc_exit (param i32)))

  ;; Memory (required by WASI)
  (memory (export "memory") 1)

  ;; _start function (WASI entry point)
  (func (export "_start")
    ;; Exit with code 42
    (call $proc_exit (i32.const 42))
  )
)
