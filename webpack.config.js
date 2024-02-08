const path = require('path');

module.exports = {
    entry: {
        bundle: './src/index.ts',
        logFileWorker: './assets/jsutils/logFileWorker.js',
        channelsIndexDbWorker: './assets/jsutils/channelsIndexedDbWorker.js',
        dmIndexedDbWorker: './assets/jsutils/dmIndexedDbWorker.js',
        ndf: './assets/jsutils/ndf.js',
        stateIndexedDbWorker: './assets/jsutils/stateIndexedDbWorker.js',
        wasm_exec: './assets/jsutils/wasm_exec.js',
    },
    devtool: 'inline-source-map',
    mode: 'development',
    output: {
        filename: '[name].js',
        path: path.resolve(__dirname, 'dist'),
        globalObject: 'this',
        library: {
            name: 'xxdk',
            type: 'umd',
        },
        umdNamedDefine: true,
        publicPath: '/dist/',
    },
    module: {
        rules: [
            {
                test: /\.tsx?$/,
                use: 'ts-loader',
                exclude: /node_modules/,
            },
            {
                test: /\.wasm$/,
                type: 'asset/resource',
                generator: {
                    filename: 'assets/wasm/[hash][ext][query]'
                }
            }
        ]
    },
    resolve: {
        extensions: ['.tsx', '.ts', 'js' ],
    },
};
