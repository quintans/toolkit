var Poller = (function () {
    function Poller(poll_url, options) {
        this.poll_url = poll_url;
        this.connected = false;
        this.versions = {};
        this.callbacks = {};
        var self = this;

        options = options || {};
        this.config = {
            timeout: options.timeout || 60000
        };

        this.onMessage = function (eventName, callback) {
            self.callbacks[eventName] = callback;
            self.versions[eventName] = 0;
            return self;
        };

        this.poll = function () {
            var poll_interval = 0;

            $.ajax(self.poll_url, {
                type: 'GET',
                dataType: 'json',
                cache: false,
                data: self.versions,
                timeout: self.config.timeout
            }).done(function (messages) {
                for (var i = 0; i < messages.length; i++) {
                    var message = messages[i];
                    if (message.version != 0) {
                        var callback = self.callbacks[message.name];
                        callback(message.data);
                        self.versions[message.name] = message.version;
                    }
                }
                if (!self.connected && self.onConnect != null) {
                    self.onConnect();
                }
                self.connected = true;
                poll_interval = 0;
            }).fail(function () {
                if (self.connected && self.onDisconnect != null) {
                    self.onConnect();
                }
                self.connected = false;
                poll_interval = 1000;
            }).always(function () {
                setTimeout(self.poll, poll_interval);
            });
        };
    }
    return Poller;
})();
