##
# Docker image for deploying the OpenGroupware iCal service.
#
# This provides an image suitable for deployment in a production environment.
#

FROM scratch
MAINTAINER Homme Zwaagstra <hrz@geodata.soton.ac.uk>

# Install the binary.
ADD ./ogo-ical /bin/

EXPOSE 8080

ENTRYPOINT ["/bin/ogo-ical"]
