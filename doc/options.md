# cspell-nix flake module options

This file is auto-generated and do not edit directly.

## _module\.args

Additional arguments passed to each module in addition to ones
like ` lib `, ` config `,
and ` pkgs `, ` modulesPath `\.

This option is also available to all submodules\. Submodules do not
inherit args from their parent module, nor do they provide args to
their parent module or sibling submodules\. The sole exception to
this is the argument ` name ` which is provided by
parent modules to a submodule and contains the attribute name
the submodule is bound to, or a unique generated name if it is
not bound to an attribute\.

Some arguments are already passed by default, of which the
following *cannot* be changed with this option:

 - ` lib `: The nixpkgs library\.

 - ` config `: The results of all options after merging the values from all modules together\.

 - ` options `: The options declared in all modules\.

 - ` specialArgs `: The ` specialArgs ` argument passed to ` evalModules `\.

 - All attributes of ` specialArgs `
   
   Whereas option values can generally depend on other option values
   thanks to laziness, this does not apply to ` imports `, which
   must be computed statically before anything else\.
   
   For this reason, callers of the module system can provide ` specialArgs `
   which are available during import resolution\.
   
   For NixOS, ` specialArgs ` includes
   ` modulesPath `, which allows you to import
   extra modules from the nixpkgs package tree without having to
   somehow make the module aware of the location of the
   ` nixpkgs ` or NixOS directories\.
   
   ```
   { modulesPath, ... }: {
     imports = [
       (modulesPath + "/profiles/minimal.nix")
     ];
   }
   ```

For NixOS, the default value for this option includes at least this argument:

 - ` pkgs `: The nixpkgs package set according to
   the ` nixpkgs.pkgs ` option\.



*Type:*
lazy attribute set of raw value

*Declared by:*
 - [\<nixpkgs/lib/modules\.nix>](https://github.com/NixOS/nixpkgs/blob//lib/modules.nix)



## programs\.envar\.enable



Whether to enable Enable envar\.



*Type:*
boolean



*Default:*
` false `



*Example:*
` true `



## programs\.envar\.enableBashIntegration



Whether to enable Bash integration\.



*Type:*
boolean



*Default:*
[](\#opt-home\.shell\.enableBashIntegration)



*Example:*
` false `



## programs\.envar\.package



The envar package to use\.



*Type:*
package



*Default:*
` <derivation envar-1> `



## programs\.envar\.settings\.execs



Scripts to execute



*Type:*
attribute set of string



*Default:*
` { } `



## programs\.envar\.settings\.vars



Environment variables to set\.



*Type:*
attribute set of list of (submodule)



*Default:*
` { } `



## programs\.envar\.settings\.vars\.\<name>\.\*\.path



Path pattern to match\.



*Type:*
string



## programs\.envar\.settings\.vars\.\<name>\.\*\.value



Value or command to bind the variable to\.



*Type:*
string or attribute set of (string or list of string)


