
// Tables, Strings, Numbers, Booleans, Struct

class Token
  id, value
end

class Tokenizer
  input, state="Initial", index=0

  // normal list
  function eof(a, b)
    .index >= .input.size
  end

  // no parens
  function consume a, b
    if .eof
      if xd
      end
      return ""
    end

    ret := .input[.index]
    .index += 1
    ret
  end

  // empty
  function tokenize()
    tokens = []
    loop
      tokens.push(Token "EOF", "")
    end
  end

  function test
    mul, ti, line
    // multiline attribute list
  end
end

module Zoo
  class Animal
    function say(x)
      println(x)
    end
  end
  class Dog < Animal
    function say(x)
      println("Bark! " + x)
    end
  end
end
