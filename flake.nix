{
  description = "balafon";

  inputs = {
    nixpkgs.url = "github:nixos/nixpkgs?ref=nixos-unstable";
  };

  outputs =
    { self, ... }@inputs:
    let
      system = "x86_64-linux";

      devpkgs = import inputs.nixpkgs {
        inherit system;
      };
    in
    {
      devShells.${system} = {
        default = devpkgs.mkShell {
          buildInputs = with devpkgs; [
            alsa-lib
            go
          ];
        };
      };
    };
}
