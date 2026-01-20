{ lib }:
rec {
  makeVarsYamlString =
    vars:
    let
      vars' = lib.concatLists (
        lib.mapAttrsToList (
          varName: patterns:
          [ "${varName}:" ]
          ++ lib.map (
            pattern:
            if lib.isAttrs pattern.value then
              [ "${pattern.path}:" ]
              ++ lib.mapAttrsToList (
                command: args:
                if lib.isList args then
                  [
                    "${command}:"
                    (lib.map (a: "- \"${a}\"") args)
                  ]
                else
                  [ "${command}: \"${args}\"" ]
              ) pattern.value
            else
              [ "${pattern.path}: \"${pattern.value}\"" ]
          ) patterns
        ) vars
      );
    in
    lib.concatLines (indent vars');
  /**
    indent gets a list of strings or a list of them and returns a list of strings.

    Example:
      > indent ["a" ["b" "c" ["d"]]]
      [
        "a"
        "  b"
        "  c"
        "    d"
      ]
  */
  indent = a: lib.concatMap (a: indent' a) a;
  indent' = a: if lib.isList a then lib.concatMap (a: lib.map (a: "  ${a}") (indent' a)) a else [ a ];
}
