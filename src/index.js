////////////////////////////////////////////////////////////////////////////////
// Copyright © 2022 xx foundation                                             //
//                                                                            //
// Use of this source code is governed by a license that can be found in the  //
// LICENSE file.                                                              //
////////////////////////////////////////////////////////////////////////////////

// import _ from 'lodash';
// import { divide } from "./api/main.go";
import wasm from './api/main.go';


const { divide } = wasm;

// function component() {
//     const element = document.createElement('div');
//
//     // Lodash, now imported by this script
//     element.innerHTML = _.join(['Hello', 'webpack'], ' ');
//
//     return element;
// }
//
// document.body.appendChild(component());

const result = await divide(6, 2);
console.log(result); // 3
