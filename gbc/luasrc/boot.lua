if type(WORKER) == "table" then
    package.path = string.format("%s/?.lua;./?.lua", WORKER.BOOT_DIR)
else
    error("ERR: NOT FOUND GLOBAL VAR 'WORKER'")
end

require("framework.init")

local GBCHandler = require("gbc.handler")

local function main()
    gbc.dump(WORKER, "WORKER")

    -- @var GBCHandler
    local handler = GBCHandler.new(WORKER.ID, WORKER.MESSAGE_CHAN, WORKER.WORKER_CHAN)
    handler:startLoop()
end

--

main()
