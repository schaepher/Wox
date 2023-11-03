default:
    @just --list --unsorted

@dev:
    just _build_hosts
    just plugins

@precommit:
    cd Wox.UI.Tauri && pnpm build && cd ..

@plugins:
    # build plugins
    #just _build_dev_nodejs_plugin Wox.Plugin.ProcessKiller ~/.wox/wox-user/plugins

    just _build_dev_nodejs_plugin Wox.Plugin.ProcessKiller ~/icloud/wox/plugins
    just _build_dev_nodejs_plugin_chatgpt Wox.Plugin.Chatgpt ~/icloud/wox/plugins

@release target:
    ENV PROD=true

    just _build_hosts
    just _build_ui

    # windows platform in hotkey doesn't need C
    if [ "{{target}}" = "windows" ]; then \
      cd Wox && CGO_ENABLED=1 GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X 'wox/util.ProdEnv=true'" -o ../Release/wox-windows-amd64.exe && cd ..; \
    elif [ "{{target}}" = "linux" ]; then \
      cd Wox && CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -ldflags "-s -w -X 'wox/util.ProdEnv=true'" -o ../Release/wox-linux && cd ..; \
    elif [ "{{target}}" = "darwin" ]; then \
      cd Wox && CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 go build -ldflags "-s -w -X 'wox/util.ProdEnv=true'" -o ../Release/wox-mac-amd64 && cd ..; \
      cd Wox && CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 go build -ldflags "-s -w -X 'wox/util.ProdEnv=true'" -o ../Release/wox-mac-arm64 && cd ..; \
    fi

@test:
    cd Wox && go test ./...

@_build_dev_nodejs_plugin pluginName directory:
    rm -rf {{directory}}/{{pluginName}}
    cd Plugins/{{pluginName}} && pnpm install && pnpm run build && cd ..
    mkdir -p {{directory}}/{{pluginName}}
    cp -r Plugins/{{pluginName}}/dist/* {{directory}}/{{pluginName}}/
    cp Plugins/{{pluginName}}/plugin.json {{directory}}/{{pluginName}}/plugin.json
    cp -r Plugins/{{pluginName}}/images {{directory}}/{{pluginName}}/images

@_build_dev_nodejs_plugin_chatgpt pluginName directory:
    rm -rf {{directory}}/{{pluginName}}
    cd Plugins/{{pluginName}} && just build && cd ..
    mkdir -p {{directory}}/{{pluginName}}
    cp -r Plugins/{{pluginName}}/dist/* {{directory}}/{{pluginName}}/
    cp Plugins/{{pluginName}}/Wox.Plugin.Chatgpt.Server/plugin.json {{directory}}/{{pluginName}}/plugin.json
    cp -r Plugins/{{pluginName}}/Wox.Plugin.Chatgpt.Server/images {{directory}}/{{pluginName}}/images

@_build_hosts:
    # build hosts
    rm -rf Wox/resource/hosts
    mkdir Wox/resource/hosts
    just _build_nodejs_host Wox/resource/hosts
    just _build_python_host Wox/resource/hosts

@_build_ui:
    cd Wox.UI.Tauri && pnpm install && pnpm release && cd ..
    cp Wox.UI.Tauri/src-tauri/target/release/wox Wox/resource/ui/wox
    chmod +x Wox/resource/ui/wox

@_build_nodejs_host directory:
    cd Wox.Plugin.Host.Nodejs && pnpm install && pnpm run build && cd ..
    mkdir -p {{directory}}
    cp Wox.Plugin.Host.Nodejs/dist/index.js {{directory}}/node-host.js

@_build_python_host directory:
    cd Wox.Plugin.Host.Python && \
    rm -rf python-host && \
    rm -rf python-host.pyz && \
    python -m pip install -r requirements.txt --target python-host && \
    cp *.py python-host && \
    python -m zipapp -p "interpreter" python-host && \
    rm -rf python-host && \
    cd ..
    mkdir -p {{directory}}
    cp Wox.Plugin.Host.Python/python-host.pyz {{directory}}/python-host.pyz