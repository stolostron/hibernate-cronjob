FROM registry.access.redhat.com/ubi8/python-38

# Add the program and the configuration file
ADD ./src/action.py .

# Install dependencies
RUN pip install --upgrade pip
RUN python -m pip install kubernetes

USER 1001