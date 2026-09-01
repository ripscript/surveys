import inspect
from pprint import pformat

def _format(obj):
    if hasattr(obj, "__dict__"):
        return f"{obj.__class__.__name__}({pformat(vars(obj))})"
    return pformat(obj)

def _build_output(args, kwargs, tag):
    lines = [f"BEGIN:DEBUG[{tag}]========================================"]
    for a in args:
        lines.append(_format(a))
    for k, v in kwargs.items():
        lines.append(f"{k} = {_format(v)}")
    lines.append(f"END:DEBUG[{tag}]==========================================")
    return "\n".join(lines)

def _get_tag(frame):
    line = frame.f_lineno
    file = frame.f_code.co_filename.split("/")[-1]
    return f"{file}:{line}"

def dprint(*args, **kwargs):
    frame = inspect.currentframe().f_back
    tag = _get_tag(frame)
    print(_build_output(args, kwargs, tag))

def dprint_file(*args, log_path="debug.log", **kwargs):
    frame = inspect.currentframe().f_back
    tag = _get_tag(frame)
    output = _build_output(args, kwargs, tag)
    with open(log_path, "a", encoding="utf-8") as f:
        f.write(output + "\n")