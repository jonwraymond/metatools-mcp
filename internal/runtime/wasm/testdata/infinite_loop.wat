;; A minimal WASI program that loops forever (for timeout testing)
(module
  ;; Memory (required by WASI)
  (memory (export "memory") 1)

  ;; _start function (WASI entry point)
  (func (export "_start")
    (loop $infinite
      (br $infinite)
    )
  )
)
