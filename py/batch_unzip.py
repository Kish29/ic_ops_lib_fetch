import os
import threading

extractable = ".zip"

os.chdir("../source_code")


# Start extract files ...
def run_cmd(cmd_str='', echo_print=1):
    """
    执行cmd命令，不显示执行过程中弹出的黑框
    备注：subprocess.run()函数会将本来打印到cmd上的内容打印到python执行界面上，所以避免了出现cmd弹出框的问题
    :param echo_print:
    :param cmd_str: 执行的cmd命令
    :return:
    """
    print(cmd_str)
    from subprocess import run
    if echo_print == 1:
        print('\nexecute command=>{}'.format(cmd_str))
    run(cmd_str, shell=True)


def extract_zip(dirname, filename):
    run_cmd("unzip {} -d {}".format(os.path.join(dirname, filename), dirname))


for root, _, files in os.walk("."):
    for file in files:
        last_for_char = file[len(file) - 4:]
        if last_for_char == extractable:
            t = threading.Thread(target=extract_zip, args=(root, file,))
            t.start()
            t.join()
