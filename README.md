# GitModuleManager
開発したライブラリを、submoduleとしてではなく、
ファイルとして管理したかったため作成  
tempディレクトリを作成しrsyncしている


依存リポジトリも記述可能で、GMMDepend.ymlにmoduleを定義すれば  
リポジトリが被っていなければ同期する

## install bin 

https://github.com/yayorozu/GitModuleManager/blob/master/gmm を DL  
PATHの通ったところへ移動

```bash
$ gmm version
```

が表示されればOK

## GitModuleFile.yml

```yaml
excludes: 
  - LICENSE,
  - README*,

root: .

modules:
  -
    path: Temp # sync path
    url: git@github.com:
    target: master # checkout target branch or tag or hash
    # excludes: 
    #  - 
```

#### excludes:
同期時に除外するオプションを記述  
※rsyncのexcludeと同じ

#### root:
同期先のディレクトリ

#### modules:
同期するモジュール一覧

#### modules:path
同期するディレクトリ

#### modules:url
clone するリポジトリ  
※httpsは未対応

#### modules:target
checkout先  
ブランチ、タグ、ハッシュが利用できる

#### modules:excludes
excludesと同じ  
リポジトリ別に設定したい際に利用する

## GMMDepend.yml
依存しているリポジトリがあればrootにおいておく  
GitModuleFile.ymlにすでに記述されいる場合はスキップする

```yaml
modules:
  -
    path: Temp # sync path
    url: git@github.com:
    target: master # checkout target branch or tag or hash
    # excludes: 
    #  - 
```

## コマンド
#### init
GitModuleFile.ymlを作成

#### initDepend
GMMDepend.ymlを作成

#### sync
同期を開始

#### resync
ルートパスの中を削除し再度syncを行う

#### clean
ルートパスの中を削除

#### cleanCache
Cloneしたリポジトリのキャッシュを削除する

#### help
コマンド一覧を表示
