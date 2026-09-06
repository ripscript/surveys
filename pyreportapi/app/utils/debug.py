import inspect
from pprint import pformat

def _format(obj, _seen=None):
    if _seen is None:
        _seen = set()

    obj_id = id(obj)
    if obj_id in _seen:
        return "<circular reference>"

    if hasattr(obj, "__dict__"):
        _seen = _seen | {obj_id}
        attrs = vars(obj)
        formatted = {
            k: _format(v, _seen)
            for k, v in attrs.items()
            if k != "_sa_instance_state"  # opsional: sembunyikan internal SQLAlchemy
        }
        return f"{obj.__class__.__name__}({pformat(formatted)})"

    if isinstance(obj, dict):
        return {k: _format(v, _seen) for k, v in obj.items()}

    if isinstance(obj, (list, tuple, set)):
        return type(obj)(_format(v, _seen) for v in obj)

    return obj

def _build_output(args, kwargs, tag):
    lines = [f"BEGIN:DEBUG[{tag}]========================================"]
    for a in args:
        result = _format(a)
        lines.append(result if isinstance(result, str) else pformat(result))
    for k, v in kwargs.items():
        result = _format(v)
        formatted = result if isinstance(result, str) else pformat(result)
        lines.append(f"{k} = {formatted}")
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