local proto = require("proto")

--- @class gbc.MessageHandler
local MessageHandler = gbc.Class("MessageHandler")
gbc.MessageHandler = MessageHandler

function MessageHandler:Constructor(id, workerChan, inputChan, outputChan)
    self.id = id
    self.workerChan = workerChan
    self.inputChan = inputChan
    self.outputChan = outputChan
end

function MessageHandler:SetIdle()
    self.workerChan:send(self.id)
end

function MessageHandler:StartLoop()
    self:SetIdle()

    local inputChan = self.inputChan
    local exit = false

    gbc.Printf("- GBCHandler %s loop start", self.id)

    while not exit do
        channel.select({ "|<-", inputChan, function(ok, msg)
            if not ok then
                -- channel is closed
                exit = true
            else
                self:ReceiveProtoMessage(msg)
                self:SetIdle()
            end
        end })
    end

    gbc.Printf("- GBCHandler %s loop end", self.id)
end

function MessageHandler:ReceiveProtoMessage(msg)
    gbc.Printf("- GBCHandler %s receive message: %s", self.id, tostring(msg))
end

function MessageHandler:SendMessage(msg)
end
