load("exec.star", "exec")

def execute(command):
    cmd = exec.new(command)
    print("Command ID:", cmd)

    ret_code = exec.run(cmd)
    if ret_code != 0:
        return "error"
    
    output  = exec.output(cmd)

    return output