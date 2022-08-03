# Wlang (WIP)
> Experiment on a language that aims to be simple and confortable, while reasonably fast.

<img src="https://i.imgur.com/kgevhFq.png" width="250px" height="150px"></img>

## What should Wlang look like?

Hopefully familiar and easy to understand, the design is currently constantly changing.
However I'm working with the examples in the test-assets folder to get the basic compiler running.

```ruby
import "io"

class Cat
  :name, :age

  def say
    io.puts(
      "Meow! My name is %name, and i'm %age years old",
      name: .name, age: .age
    )
  end
end

def main(argv)
  cat = Cat.new("Mr. Clinton", 10)
  cat.say()
end
```

The code above would then transpile to human readable C99 and be statically linked.
Memory would be managed by the runtime. (Probably refcounted).
Something like:
```c
#include "wlang/runtime.h"
#include "wlang/io.h"

typedef struct w_main_class_cat {
  WMetadata __metadata;
  WValue name;
  WValue age;
} w_main_class_cat;

WValue w_main_class_cat_say (w_main_class_cat* self) {
  wlang_io_puts(
    "Meow! My name is %name, and i'm %age years old",
    w_table_build((Wvalue[]){
      w_cstring("name"), self->name, w_cstring("age"), self->age
    }, 4)
  );
}

int main(int __argc, char** __argv) {
  WValue argv = w_argv_to_table(__argc, __argv);
  w_main_class_cat cat = {.name=w_cstring("Mr. Clinton"), .age=w_cint(10)};
  w_main_class_cat_say(cat);
}
```

There are supposed to be only a VERY small set of Primitive Types:
- Table
  > Works as both an Array and a Hash
- String
  > Text decoding/encoding, iteration and everything else you expect a string to do.
  > Maybe just use a Table of Number ? Or a struct with a table of Number and encoding
- Number
  > Variable Length Integer/Float
- Boolean
  > t/f
- Object
  > Any class instance or function

## Compiler Infrastucture Progress
- [x] Tokenizer
- [ ] Parser:
  - [x] Statement Parsing
  - [ ] Expression Parsing
  - [x] AST Generation
- [ ] Semantic analysis
- [ ] Codegen
