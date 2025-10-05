# shell.nix
let
  # Configure Nix to allow unfree packages.
  config = {
    allowUnfree = true;
  };
  pkgs = import <nixpkgs> { inherit config; };
in
pkgs.mkShell {
  name = "golang";
  
  buildInputs = with pkgs; [
    go
    gopls
    delve
    gotools
    sqlite
    # postman
  ];
  shellHook = ''
    # 设置 GOPATH 和 GOBIN
    export GOPATH=$PWD/.go
    export GOBIN=$GOPATH/bin
    export PATH=$GOBIN:$PATH
    
    # 创建必要的目录
    mkdir -p $GOPATH $GOBIN
    
    go env -w GOPROXY="https://goproxy.cn,direct"
    echo "Go development environment ready!"
    echo "Go version: $(go version)"
  '';
}