/*
 * Copyright © 2020 xx network SEZC                                           ///
 *                                                                            ///
 * Use of this source code is governed by a license that can be found in the  ///
 * LICENSE file                                                               ///
 */

async function SendE2e(ndf, recipientContactFile, myContactFileName, statePath, statePassString) {
    let enc = new TextEncoder();
    let dec = new TextDecoder();

    const output = document.getElementById("output")
    const statePass = enc.encode(statePassString);

    // Check if state exists
    if (localStorage.getItem(statePath) === null) {
        console.log('getting key ' + statePath + ' returned null; making new cmix');

        output.innerHTML += "Loading new storage\n"

        // Initialize the state
        NewCmix(ndf, statePath, statePass, '');
    }


    ////////////////////////////////////////////////////////////////////////////
    // Login to your client session                                           //
    ////////////////////////////////////////////////////////////////////////////

    // Login with the same statePath and statePass used to call NewCmix
    let netID = LoadCmix(statePath, statePass, GetDefaultCMixParams());
    console.log("LoadCmix() " + netID)

    console.log("sleep start")
    await sleep(3000)
    console.log("sleep end")

    let net = GetLoadCmix(netID)
    console.log("loaded cmix: " + net)

    // Get reception identity (automatically created if one does not exist)
    const identityStorageKey = "identityStorageKey";
    let identity;
    try {
        identity = LoadReceptionIdentity(identityStorageKey, net.GetID());
    } catch {
        // If no extant xxdk.ReceptionIdentity, generate and store a new one
        identity = net.MakeReceptionIdentity();

        StoreReceptionIdentity(identityStorageKey, identity, net.GetID());
    }

    // Print contact to console. This should probably save a file.
    const myContactFile = dec.decode(GetContactFromReceptionIdentity(identity))
    console.log("my contact file content: " + myContactFile);

    // Start file download.
    download(myContactFileName, myContactFile);

    let confirm = false;
    let confirmContact;
    let authCallbacks = {
        Confirm: function (contact, receptionId, ephemeralId, roundId) {
            confirm = true;
            confirmContact = contact
            console.log("Confirm:");
            console.log("contact: " + dec.decode(contact));
            console.log("receptionId: " + ephemeralId.toString());
            console.log("ephemeralId: " + roundId.toString());

            output.innerHTML += "Received confirmation from " + ephemeralId.toString() + "<br />"
        }
    }

    // Create an E2E client
    // Pass in auth object which controls auth callbacks for this client
    const params = GetDefaultE2EParams();
    console.log("Using E2E parameters: " + dec.decode(params));
    let e2eClient = Login(net.GetID(), authCallbacks, identity, params);


    ////////////////////////////////////////////////////////////////////////////
    // Start network threads                                                  //
    ////////////////////////////////////////////////////////////////////////////

    // Set networkFollowerTimeout to a value of your choice (seconds)
    net.StartNetworkFollower(5000);

    output.innerHTML += "Starting network follower<br />"

    // Set up a wait for the network to be connected
    let health = false
    const n = 100
    let myPromise = new Promise(async function (myResolve, myReject) {
        for (let i = 0; (health === false) && (i < n); i++) {
            await sleep(100)
        }
        if (health === true) {
            myResolve("OK");
        } else {
            myReject("timed out waiting for healthy network");
        }
    });

    // Provide a callback that will be signalled when network health status changes
    net.AddHealthCallback({
        Callback: function (healthy) {
            health = healthy;
        }
    });
    await sleep(3000)

    // Wait until connected or crash on timeout
    myPromise.then(
        function (value) {
            output.innerHTML += "Network is healthy<br />"
            console.log("network is healthy")
        },
        function (error) {
            output.innerHTML += "Network is not healthy<br />"
            // throw error;
        }
    );


    ////////////////////////////////////////////////////////////////////////////
    // Register a listener for messages                                       //
    ////////////////////////////////////////////////////////////////////////////

    let listener = {
        Hear: function (item) {
            console.log("Listener heard: " + dec.decode(item));
            output.innerHTML += "Listener heard: " + dec.decode(item) + "<br />"
        },
        Name: function () {
            return "Listener";
        }
    }

    // Listen for all types of messages using catalog.NoType
    // Listen for messages from all users using id.ZeroUser
    let zerUser = Uint8Array.from([0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0,
        0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 3]);
    e2eClient.RegisterListener(zerUser, 0, listener);

    output.innerHTML += "Registered listener<br />"

    ////////////////////////////////////////////////////////////////////////////
    // Connect with the recipient                                             //
    ////////////////////////////////////////////////////////////////////////////

    // Check that the partner exists
    if (recipientContactFile !== '') {
        let exists = false;
        output.innerHTML += "getting ID from contact<br />"
        const recipientContactID = GetIDFromContact(recipientContactFile);
        output.innerHTML += "Checking for " + btoa(recipientContactID) + "<br />"
        let partners = e2eClient.GetAllPartnerIDs();

        for (let i = 0; i < partners.length; i++) {

            if (partners[i] === recipientContactID) {
                console.log("partner " + btoa(recipientContactID) + " matches partner " + i + " " + btoa(partners[i]))
                exists = true;
                break
            }
        }

        // If the partner does not exist, send a request
        if (exists === false) {
            output.innerHTML += "Request sent to " + btoa(recipientContactID) + "<br />"
            const factList = enc.encode('[]')
            e2eClient.Request(enc.encode(recipientContactFile), factList)

            for (let i = 0; (i < 600) && (confirm === false); i++) {
                await sleep(50)
            }
            if (confirm === false) {
                output.innerHTML += "Checking for " + recipientContactID + "<br />"
                throw new Error("timed out waiting for confirmation")
            }

            const confirmContactID = GetIDFromContact(confirmContact)
            if (recipientContactID !== confirmContactID) {
                throw new Error("contact ID from confirmation " +
                    btoa(dec.decode(confirmContactID)) +
                    " does not match recipient ID " +
                    btoa(dec.decode(recipientContactID)))
            }
        }

        ////////////////////////////////////////////////////////////////////////////
        // Send a message to the recipient                                        //
        ////////////////////////////////////////////////////////////////////////////

        // Test message
        const msgBody = "If this message is sent successfully, we'll have established contact with the recipient."

        output.innerHTML += "Sending E2E message<br />"
        const paramsObj = JSON.parse(dec.decode(params))
        const e2eSendReport = e2eClient.SendE2E(2, recipientContactID, enc.encode(msgBody), enc.encode(JSON.stringify(paramsObj.Base)))

        console.log("e2e send report: " + dec.decode(e2eSendReport))
        output.innerHTML += "Send e2e: " + dec.decode(e2eSendReport) + "<br />"
    } else {
        output.innerHTML += "Partner does not exist<br />"
    }
}

function sleep(ms) {
    return new Promise(resolve => setTimeout(resolve, ms));
}

function download(filename, text) {
    let element = document.createElement('a');
    element.setAttribute('href', 'data:text/plain;charset=utf-8,' + encodeURIComponent(text));
    element.setAttribute('download', filename);

    element.style.display = 'none';
    document.body.appendChild(element);

    element.click();

    document.body.removeChild(element);
}
