local handler = gbc.class("GBCHandler")

function handler:ctor(id, messageChan, workerChan)
    self.id = id
    self.messageChan = messageChan
    self.workerChan = workerChan
end

function handler:setIdle()
    self.workerChan:send(self.id)
end

function handler:startLoop()
    self:setIdle()

    local mc = self.messageChan
    local exit = false

    print("- GBCHandler loop start")

    while not exit do
        channel.select({ "|<-", mc, function(ok, msg)
            if not ok then
                exit = true
            else
                self:ReceiveProtoMessage(msg)
                self:setIdle()
            end
        end })
    end

    print("- GBCHandler loop exit")
end

function handler:ReceiveProtoMessage(msg)
    gbc.dump(msg, "received message")
end

return handler
