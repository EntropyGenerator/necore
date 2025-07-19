# shell.nix
{ pkgs ? import <nixpkgs> {} }:
pkgs.mkShell {
  name = "golang";
  
  buildInputs = with pkgs; [
    go
    gopls
    delve
    gotools

  ];
  shellHook = ''
    # 设置 GOPATH 和 GOBIN
    export GOPATH=$PWD/.go
    export GOBIN=$GOPATH/bin
    export PATH=$GOBIN:$PATH
    
    # 创建必要的目录
    mkdir -p $GOPATH $GOBIN
    
    go env -w GOPROXY="https://repo.nju.edu.cn/go/,direct"
    echo "Go development environment ready!"
    echo "Go version: $(go version)"
  '';
}