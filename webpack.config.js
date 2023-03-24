////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

const path = require('path');

module.exports = {
    entry: './src/index.js',
    devServer: {
        static: {
            directory: path.resolve(__dirname, 'dist'),
        },
        compress: true,
        port: 3000,
    },
    output: {
        filename: 'main.js',
        path: path.resolve(__dirname, 'dist'),
    },
    mode: 'development',
    experiments: {
        topLevelAwait: true
    },
    resolve: {
        extensions: [".js", ".go"],
        fallback: {
            "fs": false,
            "os": false,
            "util": false,
            "tls": false,
            "net": false,
            "path": false,
            "zlib": false,
            "http": false,
            "https": false,
            "stream": false,
            "crypto": false,
        }
    },
    module: {
        rules: [
            {
                test: /\.go$/,
                use: [ 'golang-wasm' ]
            }
        ]
    },
    ignoreWarnings: [
        {
            module: /wasm_exec.js$/
        }
    ]
};